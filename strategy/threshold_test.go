package strategy_test

import (
	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/autoscaler/inventory"
	"github.com/notonthehighstreet/autoscaler/monitor"
	"github.com/notonthehighstreet/autoscaler/strategy"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

var mockResponse []monitor.MetricUpdate
var thresholdStrategy *strategy.Threshold

func setupThresholdTest() {
	log = logrus.WithFields(logrus.Fields{
		"manager":  "Mock",
		"strategy": "ThresholdStrategy",
	})
	config = viper.New()
	m, _ := monitor.MockNew(config, log)
	mockMonitor = m.(*monitor.MockMonitor)
	i, _ := inventory.MockNew(config, log)
	mockInventory, _ = i.(*inventory.MockInventory)
	t, _ := strategy.NewThreshold(config, mockInventory, mockMonitor, log)
	thresholdStrategy = t.(*strategy.Threshold)
	mockMonitor.On("GetUpdatedMetrics").Return(&mockResponse, nil)
}

func TestThresholdStrategy_EvaluateSingleMetric(t *testing.T) {
	setupThresholdTest()

	config.Set("thresholds.metric.name.min", 5)
	config.Set("thresholds.metric.name.max", 15)

	mockResponse = []monitor.MetricUpdate{{Name: "metric.name", CurrentReading: 0}}
	recommendation, _ := thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, strategy.SCALEDOWN)

	mockResponse = []monitor.MetricUpdate{{Name: "metric.name", CurrentReading: 10}}
	recommendation, _ = thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, strategy.HOLD)

	mockResponse = []monitor.MetricUpdate{{Name: "metric.name", CurrentReading: 20}}
	recommendation, _ = thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, strategy.SCALEUP)

	setupThresholdTest()
	config.Set("thresholds.metric.name.foo", "invalid")
	mockResponse = []monitor.MetricUpdate{{Name: "metric.name", CurrentReading: 0}}
	_, err := thresholdStrategy.Evaluate()
	assert.Error(t, err)

	config.Set("thresholds.metric.name.min", 5)
	recommendation, _ = thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, strategy.SCALEDOWN)

	mockResponse = []monitor.MetricUpdate{{Name: "metric.name", CurrentReading: 10}}
	recommendation, _ = thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, strategy.HOLD)

	config.Set("thresholds.metric.name.max", 6)
	recommendation, _ = thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, strategy.SCALEUP)
}

func TestThresholdStrategy_EvaluateSingleMetricInverted(t *testing.T) {
	setupThresholdTest()

	config.Set("thresholds.metric.name.min", 5)
	config.Set("thresholds.metric.name.max", 15)
	config.Set("thresholds.metric.name.invert_scaling", true)

	mockResponse = []monitor.MetricUpdate{{Name: "metric.name", CurrentReading: 0}}
	recommendation, _ := thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, strategy.SCALEUP)

	mockResponse = []monitor.MetricUpdate{{Name: "metric.name", CurrentReading: 10}}
	recommendation, _ = thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, strategy.HOLD)

	mockResponse = []monitor.MetricUpdate{{Name: "metric.name", CurrentReading: 20}}
	recommendation, _ = thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, strategy.SCALEDOWN)
}

func TestThresholdStrategy_EvaluateMultipleMetrics(t *testing.T) {
	setupThresholdTest()

	config.Set("thresholds.metric.one.min", 5)
	config.Set("thresholds.metric.one.max", 15)
	config.Set("thresholds.metric.two.min", 5)
	config.Set("thresholds.metric.two.max", 15)

	mockResponse = []monitor.MetricUpdate{
		{Name: "metric.one", CurrentReading: 0},
		{Name: "metric.two", CurrentReading: 0},
	}
	recommendation, _ := thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, strategy.SCALEDOWN)

	mockResponse = []monitor.MetricUpdate{
		{Name: "metric.one", CurrentReading: 0},
		{Name: "metric.two", CurrentReading: 10},
	}
	recommendation, _ = thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, strategy.HOLD)

	mockResponse = []monitor.MetricUpdate{
		{Name: "metric.one", CurrentReading: 0},
		{Name: "metric.two", CurrentReading: 20},
	}
	recommendation, _ = thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, strategy.SCALEUP)

	mockResponse = []monitor.MetricUpdate{
		{Name: "metric.one", CurrentReading: 10},
		{Name: "metric.two", CurrentReading: 0},
	}
	recommendation, _ = thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, strategy.HOLD)

	mockResponse = []monitor.MetricUpdate{
		{Name: "metric.one", CurrentReading: 20},
		{Name: "metric.two", CurrentReading: 0},
	}
	recommendation, _ = thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, strategy.SCALEUP)

}
