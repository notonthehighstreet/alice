package mesos

import (
	"github.com/Sirupsen/logrus"
	"github.com/andygrunwald/megos"
	"github.com/notonthehighstreet/autoscaler/manager/monitor"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"net/url"
)

type MesosMonitor struct {
	log    *logrus.Entry
	Client MesosClient
	config *viper.Viper
}

type MesosStats struct {
	CPUUsed      float64
	CPUAvailable float64
	CPUPercent   float64

	MemUsed      float64
	MemAvailable float64
	MemPercent   float64
}

type MesosClient interface {
	GetStateFromLeader() (*megos.State, error)
	DetermineLeader() (*megos.Pid, error)
}

const defaultMesosMaster = "http://mesos.service.consul:5050/state"

// NewMesosMaster initialises any new Mesos master. We will use this master to determine the leader of the cluster.
func New(config *viper.Viper, log *logrus.Entry) (monitor.Monitor, error) {
	config.SetDefault("endpoint", defaultMesosMaster)
	u, err := url.Parse(config.GetString("endpoint"))
	if err != nil {
		return nil, errors.Wrap(err, "Can't create mesos monitor")
	}
	mesos := megos.NewClient([]*url.URL{u}, nil)
	return &MesosMonitor{log: log, Client: mesos, config: config}, nil
}

func (m *MesosMonitor) GetUpdatedMetrics(names []string) (*[]monitor.MetricUpdate, error) {
	response := make([]monitor.MetricUpdate, len(names))
	stats, err := m.Stats()
	if err != nil {
		return nil, err
	}
	for i, name := range names {
		response[i].Name = name
		switch name {
		case "mesos.cluster.cpu_percent":
			response[i].CurrentReading = float64(stats.CPUPercent * 100)
		case "mesos.cluster.mem_percent":
			response[i].CurrentReading = float64(stats.MemPercent * 100)
		default:
			return &response, errors.Errorf("Unknown mesos metric: %s", name)
		}
	}
	return &response, nil
}

func (m *MesosMonitor) Stats() (*MesosStats, error) {
	m.Client.DetermineLeader()
	state, err := m.Client.GetStateFromLeader()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting Mesos stats")
	}

	stats := &MesosStats{}

	for _, slave := range state.Slaves {
		stats.CPUAvailable += slave.UnreservedResources.CPUs
		stats.MemAvailable += slave.UnreservedResources.Mem
		stats.CPUUsed += slave.UsedResources.CPUs
		stats.MemUsed += slave.UsedResources.Mem
	}
	stats.CPUPercent = stats.CPUUsed / stats.CPUAvailable
	stats.MemPercent = stats.MemUsed / stats.MemAvailable

	return stats, nil
}
