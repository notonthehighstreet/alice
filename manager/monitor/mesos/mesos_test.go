package mesos_test

import (
	"github.com/andygrunwald/megos"
	"github.com/op/go-logging"
	"github.com/stretchr/testify/assert"

	"github.com/notonthehighstreet/autoscaler/manager/monitor/mesos"
	"testing"
)

var log = logging.MustGetLogger("autoscaler")
var mockMesosClient mesos.MockMesosClient
var state megos.State

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
}

func TestCalculatesStatistics(t *testing.T) {
	setupTest()
	monitor := mesos.New(log, &mockMesosClient)
	stats := monitor.Stats()
	assert.Equal(t, 0.25, stats.CPUPercent)
	assert.Equal(t, 0.5, stats.MemPercent)
}

func TestMesosMaster_GetUpdatedMetrics(t *testing.T) {
	setupTest()
	monitor := mesos.New(log, &mockMesosClient)
	metrics, err := monitor.GetUpdatedMetrics([]string{"mesos.cluster.cpu.percent_used"})
	assert.Nil(t, err)
	assert.Equal(t, 25, (*metrics)[0].CurrentReading)
	metrics, err = monitor.GetUpdatedMetrics([]string{"mesos.cluster.mem.percent_used"})
	assert.Nil(t, err)
	assert.Equal(t, 50, (*metrics)[0].CurrentReading)
	metrics, err = monitor.GetUpdatedMetrics([]string{"invalid.metric.name"})
	assert.NotNil(t, err)
}
