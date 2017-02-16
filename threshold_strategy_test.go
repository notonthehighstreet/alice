package autoscaler_test

import (
	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/autoscaler"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

var mockResponse []autoscaler.MetricUpdate
var thresholdStrategy *autoscaler.ThresholdStrategy

func setupThresholdStrategyTest() {
	log = logrus.WithFields(logrus.Fields{
		"manager":  "Mock",
		"strategy": "ThresholdStrategy",
	})
	config = viper.New()
	m, _ := autoscaler.NewMockMonitor(config, log)
	mockMonitor = m.(*autoscaler.MockMonitor)
	i, _ := autoscaler.NewMockInventory(config, log)
	mockInventory, _ = i.(*autoscaler.MockInventory)
	t, _ := autoscaler.NewThresholdStrategy(config, mockInventory, mockMonitor, log)
	thresholdStrategy = t.(*autoscaler.ThresholdStrategy)
	mockMonitor.On("GetUpdatedMetrics").Return(&mockResponse, nil)
}

func TestThresholdStrategy_EvaluateSingleMetric(t *testing.T) {
	setupThresholdStrategyTest()

	config.Set("thresholds.metric.name.min", 5)
	config.Set("thresholds.metric.name.max", 15)

	mockResponse = []autoscaler.MetricUpdate{{Name: "metric.name", CurrentReading: 0}}
	recommendation, _ := thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, autoscaler.SCALEDOWN)

	mockResponse = []autoscaler.MetricUpdate{{Name: "metric.name", CurrentReading: 10}}
	recommendation, _ = thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, autoscaler.HOLD)

	mockResponse = []autoscaler.MetricUpdate{{Name: "metric.name", CurrentReading: 20}}
	recommendation, _ = thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, autoscaler.SCALEUP)

	setupThresholdStrategyTest()
	config.Set("thresholds.metric.name.foo", "invalid")
	mockResponse = []autoscaler.MetricUpdate{{Name: "metric.name", CurrentReading: 0}}
	_, err := thresholdStrategy.Evaluate()
	assert.Error(t, err)

	config.Set("thresholds.metric.name.min", 5)
	recommendation, _ = thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, autoscaler.SCALEDOWN)

	mockResponse = []autoscaler.MetricUpdate{{Name: "metric.name", CurrentReading: 10}}
	recommendation, _ = thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, autoscaler.HOLD)

	config.Set("thresholds.metric.name.max", 6)
	recommendation, _ = thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, autoscaler.SCALEUP)
}

func TestThresholdStrategy_EvaluateSingleMetricInverted(t *testing.T) {
	setupThresholdStrategyTest()

	config.Set("thresholds.metric.name.min", 5)
	config.Set("thresholds.metric.name.max", 15)
	config.Set("thresholds.metric.name.invert_scaling", true)

	mockResponse = []autoscaler.MetricUpdate{{Name: "metric.name", CurrentReading: 0}}
	recommendation, _ := thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, autoscaler.SCALEUP)

	mockResponse = []autoscaler.MetricUpdate{{Name: "metric.name", CurrentReading: 10}}
	recommendation, _ = thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, autoscaler.HOLD)

	mockResponse = []autoscaler.MetricUpdate{{Name: "metric.name", CurrentReading: 20}}
	recommendation, _ = thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, autoscaler.SCALEDOWN)
}

func TestThresholdStrategy_EvaluateMultipleMetrics(t *testing.T) {
	setupThresholdStrategyTest()

	config.Set("thresholds.metric.one.min", 5)
	config.Set("thresholds.metric.one.max", 15)
	config.Set("thresholds.metric.two.min", 5)
	config.Set("thresholds.metric.two.max", 15)

	mockResponse = []autoscaler.MetricUpdate{
		{Name: "metric.one", CurrentReading: 0},
		{Name: "metric.two", CurrentReading: 0},
	}
	recommendation, _ := thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, autoscaler.SCALEDOWN)

	mockResponse = []autoscaler.MetricUpdate{
		{Name: "metric.one", CurrentReading: 0},
		{Name: "metric.two", CurrentReading: 10},
	}
	recommendation, _ = thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, autoscaler.HOLD)

	mockResponse = []autoscaler.MetricUpdate{
		{Name: "metric.one", CurrentReading: 0},
		{Name: "metric.two", CurrentReading: 20},
	}
	recommendation, _ = thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, autoscaler.SCALEUP)

	mockResponse = []autoscaler.MetricUpdate{
		{Name: "metric.one", CurrentReading: 10},
		{Name: "metric.two", CurrentReading: 0},
	}
	recommendation, _ = thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, autoscaler.HOLD)

	mockResponse = []autoscaler.MetricUpdate{
		{Name: "metric.one", CurrentReading: 20},
		{Name: "metric.two", CurrentReading: 0},
	}
	recommendation, _ = thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, autoscaler.SCALEUP)

}
