package mesos_test

import (
	"github.com/andygrunwald/megos"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/notonthehighstreet/autoscaler/manager/monitor/mesos"
	"github.com/spf13/viper"
	"testing"
)

var log = logrus.WithFields(logrus.Fields{
	"manager": "Mock",
	"monitor": "MesosMonitor",
})
var mockMesosClient mesos.MockMesosClient
var state megos.State
var monitor *mesos.MesosMonitor

func setupTest() {
	state.Slaves = []megos.Slave{
		{
			UnreservedResources: megos.Resources{CPUs: 1.0, Disk: 1.1, Mem: 1.2},
			UsedResources:       megos.Resources{CPUs: 0.5, Disk: 0.0, Mem: 1.1},
		},
		{
			UnreservedResources: megos.Resources{CPUs: 2.0, Disk: 2.1, Mem: 2.2},
			UsedResources:       megos.Resources{CPUs: 0.5, Disk: 0.0, Mem: 1.1},
		},
		{
			UnreservedResources: megos.Resources{CPUs: 3.0, Disk: 3.1, Mem: 3.2},
			UsedResources:       megos.Resources{CPUs: 0.5, Disk: 0.0, Mem: 1.1},
		},
	}
	mockMesosClient.On("GetStateFromLeader").Return(state, nil)
	mockMesosClient.On("DetermineLeader").Return(megos.Pid{}, nil)
	monitor = mesos.New(viper.New(), log).(*mesos.MesosMonitor)
	monitor.Client = &mockMesosClient
}

func TestCalculatesStatistics(t *testing.T) {
	setupTest()
	stats := monitor.Stats()
	assert.Equal(t, 0.25, stats.CPUPercent)
	assert.Equal(t, 0.5, stats.MemPercent)
}

func TestMesosMaster_GetUpdatedMetrics(t *testing.T) {
	setupTest()
	metrics, err := monitor.GetUpdatedMetrics([]string{"mesos.cluster.cpu.percent_used"})
	assert.Nil(t, err)
	assert.Equal(t, 25, (*metrics)[0].CurrentReading)
	metrics, err = monitor.GetUpdatedMetrics([]string{"mesos.cluster.mem.percent_used"})
	assert.Nil(t, err)
	assert.Equal(t, 50, (*metrics)[0].CurrentReading)
	metrics, err = monitor.GetUpdatedMetrics([]string{"invalid.metric.name"})
	assert.NotNil(t, err)
}
