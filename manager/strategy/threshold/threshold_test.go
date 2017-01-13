package threshold_test

import (
	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/autoscaler/manager/inventory"
	"github.com/notonthehighstreet/autoscaler/manager/monitor"
	"github.com/notonthehighstreet/autoscaler/manager/strategy"
	"github.com/notonthehighstreet/autoscaler/manager/strategy/threshold"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

var config *viper.Viper
var log = logrus.WithFields(logrus.Fields{
	"manager":  "Mock",
	"strategy": "ThresholdStrategy",
})
var mockInventory inventory.MockInventory
var mockMonitor monitor.MockMonitor
var mockResponse []monitor.MetricUpdate
var thr *threshold.ThresholdStrategy

func setupTest() {
	mockMonitor.On("GetUpdatedMetrics").Return(&mockResponse, nil)
	config = viper.New()

	t, _ := threshold.New(config, &mockInventory, &mockMonitor, log)
	thr = t.(*threshold.ThresholdStrategy)
}

func TestThresholdStrategy_EvaluateSingleMetric(t *testing.T) {
	setupTest()

	config.Set("thresholds.metric.name.min", 5)
	config.Set("thresholds.metric.name.max", 15)

	mockResponse = []monitor.MetricUpdate{{Name: "metric.name", CurrentReading: 0}}
	recommendation, _ := thr.Evaluate()
	assert.Equal(t, *recommendation, strategy.SCALEDOWN)

	mockResponse = []monitor.MetricUpdate{{Name: "metric.name", CurrentReading: 10}}
	recommendation, _ = thr.Evaluate()
	assert.Equal(t, *recommendation, strategy.HOLD)

	mockResponse = []monitor.MetricUpdate{{Name: "metric.name", CurrentReading: 20}}
	recommendation, _ = thr.Evaluate()
	assert.Equal(t, *recommendation, strategy.SCALEUP)
}

func TestThresholdStrategy_EvaluateSingleMetricInverted(t *testing.T) {
	setupTest()

	config.Set("thresholds.metric.name.min", 5)
	config.Set("thresholds.metric.name.max", 15)
	config.Set("thresholds.metric.name.invert_scaling", true)

	mockResponse = []monitor.MetricUpdate{{Name: "metric.name", CurrentReading: 0}}
	recommendation, _ := thr.Evaluate()
	assert.Equal(t, *recommendation, strategy.SCALEUP)

	mockResponse = []monitor.MetricUpdate{{Name: "metric.name", CurrentReading: 10}}
	recommendation, _ = thr.Evaluate()
	assert.Equal(t, *recommendation, strategy.HOLD)

	mockResponse = []monitor.MetricUpdate{{Name: "metric.name", CurrentReading: 20}}
	recommendation, _ = thr.Evaluate()
	assert.Equal(t, *recommendation, strategy.SCALEDOWN)
}

func TestThresholdStrategy_EvaluateMultipleMetrics(t *testing.T) {
	setupTest()

	config.Set("thresholds.metric.one.min", 5)
	config.Set("thresholds.metric.one.max", 15)
	config.Set("thresholds.metric.two.min", 5)
	config.Set("thresholds.metric.two.max", 15)

	mockResponse = []monitor.MetricUpdate{
		{Name: "metric.one", CurrentReading: 0},
		{Name: "metric.two", CurrentReading: 0},
	}
	recommendation, _ := thr.Evaluate()
	assert.Equal(t, *recommendation, strategy.SCALEDOWN)

	mockResponse = []monitor.MetricUpdate{
		{Name: "metric.one", CurrentReading: 0},
		{Name: "metric.two", CurrentReading: 10},
	}
	recommendation, _ = thr.Evaluate()
	assert.Equal(t, *recommendation, strategy.HOLD)

	mockResponse = []monitor.MetricUpdate{
		{Name: "metric.one", CurrentReading: 0},
		{Name: "metric.two", CurrentReading: 20},
	}
	recommendation, _ = thr.Evaluate()
	assert.Equal(t, *recommendation, strategy.SCALEUP)

	mockResponse = []monitor.MetricUpdate{
		{Name: "metric.one", CurrentReading: 10},
		{Name: "metric.two", CurrentReading: 0},
	}
	recommendation, _ = thr.Evaluate()
	assert.Equal(t, *recommendation, strategy.HOLD)

	mockResponse = []monitor.MetricUpdate{
		{Name: "metric.one", CurrentReading: 20},
		{Name: "metric.two", CurrentReading: 0},
	}
	recommendation, _ = thr.Evaluate()
	assert.Equal(t, *recommendation, strategy.SCALEUP)

}
