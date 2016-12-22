package mesos_test

import (
	"github.com/andygrunwald/megos"
	"github.com/op/go-logging"
	"github.com/stretchr/testify/assert"

	"github.com/notonthehighstreet/autoscaler/mesos"

	"testing"
)

var log = logging.MustGetLogger("autoscaler")
var mockMesosClient mesos.MockMesosClient
var state megos.State

func setupMesosTest() {
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
}

func TestReturnsNoError(t *testing.T) {
	setupMesosTest()
	mockMesosClient.On("GetStateFromLeader").Return(state, nil)
	master := mesos.NewMesosMaster(log, &mockMesosClient)
	_, err := master.Stats()
	assert.Equal(t, nil, err)
}

func TestCalculatesStatistics(t *testing.T) {
	setupMesosTest()
	mockMesosClient.On("GetStateFromLeader").Return(state, nil)
	master := mesos.NewMesosMaster(log, &mockMesosClient)
	stats, _ := master.Stats()
	assert.Equal(t, 0.25, stats.CPUPercent)
	assert.Equal(t, 0.5, stats.MemPercent)
}
