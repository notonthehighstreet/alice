package mesos

import (
	"errors"
	"github.com/andygrunwald/megos"
	"github.com/notonthehighstreet/autoscaler/manager/monitor"
	"github.com/sirupsen/logrus"
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

// NewMesosMaster initialises any new Mesos master. We will use this master to determine the leader of the cluster.
func New(config *viper.Viper, log *logrus.Entry) *MesosMonitor {
	config.SetDefault("endpoint", "http://mesos.service.consul:5050/state")
	u, err := url.Parse(config.GetString("endpoint"))
	if err != nil {
		log.Fatalf("Can't create mesos monitor: %v", err)
	}
	mesos := megos.NewClient([]*url.URL{u}, nil)
	return &MesosMonitor{log: log, Client: mesos, config: config}
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
	m.Client.DetermineLeader()
	state, err := m.Client.GetStateFromLeader()
	if err != nil {
		m.log.Fatalf("Error getting mesos stats: %v", err)
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
