package monitor

import (
	"github.com/Sirupsen/logrus"
	"github.com/andygrunwald/megos"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"net/url"
)

type Mesos struct {
	log    *logrus.Entry
	Client MesosClient
	config *viper.Viper
}

type MesosStats struct {
	Metrics map[string]float64
}

func (s *MesosStats) updateMinMax(name string, number float64) {
	if min, ok := s.Metrics[name+".min"]; ok {
		if min > number {
			s.Metrics[name+".min"] = number
		}
	} else {
		s.Metrics[name+".min"] = number
	}
	if max, ok := s.Metrics[name+".max"]; ok {
		if max < number {
			s.Metrics[name+".max"] = number
		}
	} else {
		s.Metrics[name+".max"] = number
	}
}

// Add a float64 value to a metric. Takes care of floating point rounding issues by converting to integers and back.
func (s *MesosStats) add(name string, number float64) {
	if val, ok := s.Metrics[name]; ok {
		a := int(val * 10) // Mesos resource numbers are to a max of 1 decimal place
		b := int(number * 10)
		s.Metrics[name] = float64(a+b) / 10
	} else {
		s.Metrics[name] = number
	}
}

type MesosClient interface {
	GetStateFromLeader() (*megos.State, error)
	DetermineLeader() (*megos.Pid, error)
}

const defaultMesosMaster = "http://mesos.service.consul:5050/state"

func NewMesos(config *viper.Viper, log *logrus.Entry) (Monitor, error) {
	config.SetDefault("endpoint", defaultMesosMaster)
	u, err := url.Parse(config.GetString("endpoint"))
	if err != nil {
		return nil, errors.Wrap(err, "Can't create mesos monitor")
	}
	mesos := megos.NewClient([]*url.URL{u}, nil)
	return &Mesos{log: log, Client: mesos, config: config}, nil
}

func (m *Mesos) GetUpdatedMetrics(names []string) (*[]MetricUpdate, error) {
	response := make([]MetricUpdate, len(names))
	stats, err := m.Stats()
	if err != nil {
		return nil, err
	}
	for i, name := range names {
		response[i].Name = name
		if val, ok := stats.Metrics[name]; ok {
			response[i].CurrentReading = val
		} else {
			return nil, errors.Errorf("Unknown mesos metric: %s", name)
		}
	}
	return &response, nil
}

func (m *Mesos) Stats() (*MesosStats, error) {
	m.Client.DetermineLeader()
	state, err := m.Client.GetStateFromLeader()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting Mesos stats")
	}

	stats := &MesosStats{}
	stats.Metrics = make(map[string]float64)

	for _, slave := range state.Slaves {
		stats.add("mesos.cluster.cpu_total", slave.UnreservedResources.CPUs)
		stats.add("mesos.cluster.cpu_used", slave.UsedResources.CPUs)
		stats.add("mesos.cluster.mem_total", slave.UnreservedResources.Mem)
		stats.add("mesos.cluster.mem_used", slave.UsedResources.Mem)

		stats.updateMinMax("mesos.slave.cpu_free", slave.UnreservedResources.CPUs-slave.UsedResources.CPUs)
		stats.updateMinMax("mesos.slave.cpu_used", slave.UsedResources.CPUs)
		stats.updateMinMax("mesos.slave.cpu_percent", (slave.UsedResources.CPUs*10)*100/(slave.UnreservedResources.CPUs*10))
		stats.updateMinMax("mesos.slave.mem_free", slave.UnreservedResources.Mem-slave.UsedResources.Mem)
		stats.updateMinMax("mesos.slave.mem_used", slave.UsedResources.Mem)
		stats.updateMinMax("mesos.slave.mem_percent", (slave.UsedResources.Mem*10)*100/(slave.UnreservedResources.Mem*10))
	}
	stats.Metrics["mesos.cluster.cpu_percent"] = stats.Metrics["mesos.cluster.cpu_used"] * 100 / stats.Metrics["mesos.cluster.cpu_total"]
	stats.Metrics["mesos.cluster.cpu_free"] = stats.Metrics["mesos.cluster.cpu_total"] - stats.Metrics["mesos.cluster.cpu_used"]
	stats.Metrics["mesos.cluster.mem_percent"] = stats.Metrics["mesos.cluster.mem_used"] * 100 / stats.Metrics["mesos.cluster.mem_total"]
	stats.Metrics["mesos.cluster.mem_free"] = stats.Metrics["mesos.cluster.mem_total"] - stats.Metrics["mesos.cluster.mem_used"]

	stats.Metrics["mesos.slave.cpu_free.avg"] = stats.Metrics["mesos.cluster.cpu_free"] / float64(len(state.Slaves))
	stats.Metrics["mesos.slave.cpu_used.avg"] = stats.Metrics["mesos.cluster.cpu_used"] / float64(len(state.Slaves))
	stats.Metrics["mesos.slave.mem_free.avg"] = stats.Metrics["mesos.cluster.mem_free"] / float64(len(state.Slaves))
	stats.Metrics["mesos.slave.mem_used.avg"] = stats.Metrics["mesos.cluster.mem_used"] / float64(len(state.Slaves))

	return stats, nil
}
