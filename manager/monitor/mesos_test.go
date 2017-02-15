package monitor_test

import (
	"github.com/Sirupsen/logrus"
	"github.com/andygrunwald/megos"
	"github.com/stretchr/testify/assert"

	"github.com/notonthehighstreet/autoscaler/manager/monitor"
	"github.com/spf13/viper"
	"testing"
)

var mockMesosClient monitor.MockMesosClient
var state megos.State
var mesosMon *monitor.Mesos

func setupMesosTest() {
	log = logrus.WithFields(logrus.Fields{
		"manager": "Mock",
		"monitor": "MesosMonitor",
	})
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
	m, _ := monitor.NewMesos(viper.New(), log)
	mesosMon = m.(*monitor.Mesos)
	mesosMon.Client = &mockMesosClient
}

func TestCalculatesStatistics(t *testing.T) {
	setupMesosTest()
	stats, err := mesosMon.Stats()
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, float64(25), stats.Metrics["mesos.cluster.cpu_percent"])
	assert.Equal(t, float64(50), stats.Metrics["mesos.cluster.mem_percent"])
	assert.Equal(t, float64(3.3), stats.Metrics["mesos.cluster.mem_free"])
	assert.Equal(t, float64(3.3), stats.Metrics["mesos.cluster.mem_used"])

	assert.Equal(t, float64(50), stats.Metrics["mesos.slave.cpu_percent.max"])
	assert.Equal(t, float64(34.375), stats.Metrics["mesos.slave.mem_percent.min"])

	assert.Equal(t, float64(2.1), stats.Metrics["mesos.slave.mem_free.max"])
}

func TestMesosMaster_GetUpdatedMetrics(t *testing.T) {
	setupMesosTest()
	_, err := mesosMon.GetUpdatedMetrics([]string{"invalid.metric.name"})
	assert.NotNil(t, err)
}