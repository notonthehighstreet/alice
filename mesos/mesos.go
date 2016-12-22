package mesos

import (
	"net/url"

	"github.com/andygrunwald/megos"
	"github.com/op/go-logging"
)

type MesosMaster struct {
	URL    string
	logger *logging.Logger
}

type MesosStats struct {
	CPUUsed      float64
	CPUAvailable float64
	CPUPercent   float64

	MemUsed      float64
	MemAvailable float64
	MemPercent   float64
}

// MesosURL refers to the internal path to the Mesos master on this instance.
const MesosURL = "http://mesos.service.consul:5050/state"

// NewMesosMaster initialises any new Mesos master. We will use this master to determine the leader of the cluster.
func NewMesosMaster(logger *logging.Logger) *MesosMaster {
	return &MesosMaster{URL: MesosURL, logger: logger}
}

func (m *MesosMaster) Stats() (*MesosStats, error) {
	mesosNode, err := url.Parse(m.URL)
	if err != nil {
		return nil, err
	}

	mesos := megos.NewClient([]*url.URL{mesosNode}, nil)
	mesos.DetermineLeader()

	state, err := mesos.GetStateFromLeader()
	if err != nil {
		return nil, err
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

// LogUsage pipes the current state of the resources available to this Mesos master.
func (s *MesosStats) LogUsage(log *logging.Logger) {
	log.Infof("mesos: CPUs used: %.2f of %.2f", s.CPUUsed, s.CPUAvailable)
	log.Infof("mesos: Memory used: %.2f of %.2f", s.MemUsed, s.MemAvailable)
	log.Infof("mesos: CPU usage: %.2f%%", s.CPUPercent*100)
	log.Infof("mesos: Memory usage: %.2f%%", s.MemPercent*100)
}
