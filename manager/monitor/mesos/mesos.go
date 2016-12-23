package mesos

import (
	"errors"
	"github.com/andygrunwald/megos"
	"github.com/notonthehighstreet/autoscaler/manager/monitor"
	"github.com/op/go-logging"
	"net/url"
)

type MesosMonitor struct {
	logger *logging.Logger
	client MesosClient
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
}

func NewMesosClient(URL string) (MesosClient, error) {
	mesosNode, err := url.Parse(URL)
	if err != nil {
		return nil, err
	}
	mesos := megos.NewClient([]*url.URL{mesosNode}, nil)
	mesos.DetermineLeader()
	return mesos, nil
}

// NewMesosMaster initialises any new Mesos master. We will use this master to determine the leader of the cluster.
func New(logger *logging.Logger, mesos MesosClient) *MesosMonitor {
	return &MesosMonitor{logger: logger, client: mesos}
}

func (m *MesosMonitor) GetUpdatedMetrics(names []string) (*[]monitor.MetricUpdate, error) {
	response := make([]monitor.MetricUpdate, len(names))
	stats := m.Stats()
	for i, name := range names {
		response[i].Name = name
		switch name {
		case "mesos.cluster.cpu.percent_used":
			response[i].CurrentReading = int(stats.CPUPercent * 100)
		case "mesos.cluster.mem.percent_used":
			response[i].CurrentReading = int(stats.MemPercent * 100)
		default:
			return &response, errors.New("Unknown mesos metric: " + name)
		}
	}
	return &response, nil
}

func (m *MesosMonitor) Stats() *MesosStats {
	state, err := m.client.GetStateFromLeader()
	if err != nil {
		m.logger.Fatalf("Error getting mesos stats: %v", err)
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

	return stats
}

// LogUsage pipes the current state of the resources available to this Mesos master.
func (s *MesosStats) LogUsage(log *logging.Logger) {
	log.Infof("mesos: CPUs used: %.2f of %.2f", s.CPUUsed, s.CPUAvailable)
	log.Infof("mesos: Memory used: %.2f of %.2f", s.MemUsed, s.MemAvailable)
	log.Infof("mesos: CPU usage: %.2f%%", s.CPUPercent*100)
	log.Infof("mesos: Memory usage: %.2f%%", s.MemPercent*100)
}
