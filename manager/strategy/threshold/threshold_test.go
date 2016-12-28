package threshold_test

import (
	"github.com/notonthehighstreet/autoscaler/manager/inventory"
	"github.com/notonthehighstreet/autoscaler/manager/monitor"
	"github.com/notonthehighstreet/autoscaler/manager/strategy"
	"github.com/notonthehighstreet/autoscaler/manager/strategy/threshold"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

var log = logrus.WithFields(logrus.Fields{
	"manager":  "Mock",
	"strategy": "ThresholdStrategy",
})
var mockInventory inventory.MockInventory
var mockMonitor monitor.MockMonitor
var metricNames []string
var metricUpdates []monitor.MetricUpdate

func setupTest() {
	metricNames = []string{"cpu_percent", "mem_percent", "disk_percent"}
	for i, name := range metricNames {
		metricUpdates = append(metricUpdates, monitor.MetricUpdate{Name: name, CurrentReading: i * 40}) // 0, 40, 80
	}
	mockMonitor.On("GetUpdatedMetrics").Return(&metricUpdates, nil)
}
func TestThresholdStrategy_Evaluate(t *testing.T) {
	setupTest()
	thresholds := map[string][2]int{
		"cpu_percent":  [2]int{30, 70}, // SCALEDOWN
		"mem_percent":  [2]int{30, 70}, // HOLD
		"disk_percent": [2]int{30, 70}, // SCALEUP
	}
	s := threshold.New(thresholds, &mockInventory, &mockMonitor, log)
	recommendation, error := s.Evaluate()
	assert.Nil(t, error)
	assert.Equal(t, recommendation, strategy.SCALEUP) // because of disk_percent

	s.Thresholds["disk_percent"] = [2]int{100, 100}
	recommendation, error = s.Evaluate()
	assert.Nil(t, error)
	assert.Equal(t, recommendation, strategy.HOLD) // because of mem_percent

	s.Thresholds["mem_percent"] = [2]int{100, 100}
	recommendation, error = s.Evaluate()
	assert.Nil(t, error)
	assert.Equal(t, recommendation, strategy.SCALEDOWN) // because of all
}
