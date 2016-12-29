package threshold_test

import (
	"github.com/notonthehighstreet/autoscaler/manager/inventory"
	"github.com/notonthehighstreet/autoscaler/manager/monitor"
	"github.com/notonthehighstreet/autoscaler/manager/strategy"
	"github.com/notonthehighstreet/autoscaler/manager/strategy/threshold"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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
var thr *threshold.ThresholdStrategy

func setupTest() {
	metricNames = []string{"cpu_percent", "mem_percent", "disk_percent"}
	for i, name := range metricNames {
		metricUpdates = append(metricUpdates, monitor.MetricUpdate{Name: name, CurrentReading: i * 40}) // 0, 40, 80
	}
	mockMonitor.On("GetUpdatedMetrics").Return(&metricUpdates, nil)
	config := viper.New()
	config.Set("thresholds.cpu_percent.min", 30)
	config.Set("thresholds.cpu_percent.max", 70)
	config.Set("thresholds.mem_percent.min", 30)
	config.Set("thresholds.mem_percent.max", 70)
	config.Set("thresholds.disk_percent.min", 30)
	config.Set("thresholds.disk_percent.max", 70)
	thr = threshold.New(config, &mockInventory, &mockMonitor, log)
}
func TestThresholdStrategy_Evaluate(t *testing.T) {
	setupTest()
	recommendation, error := thr.Evaluate()
	assert.Nil(t, error)
	assert.Equal(t, recommendation, strategy.SCALEUP) // because of disk_percent

	thr.Config.Set("thresholds.disk_percent.min", 100)
	thr.Config.Set("thresholds.disk_percent.max", 100)
	recommendation, error = thr.Evaluate()
	assert.Nil(t, error)
	assert.Equal(t, recommendation, strategy.HOLD) // because of mem_percent

	thr.Config.Set("thresholds.mem_percent.min", 100)
	thr.Config.Set("thresholds.mem_percent.max", 100)
	recommendation, error = thr.Evaluate()
	assert.Nil(t, error)
	assert.Equal(t, recommendation, strategy.SCALEDOWN) // because of all
}
