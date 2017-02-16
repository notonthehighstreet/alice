package alice_test

import (
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/alice"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var mockResponse []alice.MetricUpdate
var thresholdStrategy *alice.ThresholdStrategy

func setupThresholdStrategyTest() {
	log = logrus.WithFields(logrus.Fields{
		"manager":  "Mock",
		"strategy": "ThresholdStrategy",
	})
	config = viper.New()
	m, _ := alice.NewMockMonitor(config, log)
	mockMonitor = m.(*alice.MockMonitor)
	i, _ := alice.NewMockInventory(config, log)
	mockInventory, _ = i.(*alice.MockInventory)
	t, _ := alice.NewThresholdStrategy(config, mockInventory, mockMonitor, log)
	thresholdStrategy = t.(*alice.ThresholdStrategy)
	mockMonitor.On("GetUpdatedMetrics").Return(&mockResponse, nil)
}

func TestThresholdStrategy_EvaluateSingleMetric(t *testing.T) {
	setupThresholdStrategyTest()

	config.Set("thresholds.metric.name.min", 5)
	config.Set("thresholds.metric.name.max", 15)

	mockResponse = []alice.MetricUpdate{{Name: "metric.name", CurrentReading: 0}}
	recommendation, _ := thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, alice.SCALEDOWN)

	mockResponse = []alice.MetricUpdate{{Name: "metric.name", CurrentReading: 10}}
	recommendation, _ = thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, alice.HOLD)

	mockResponse = []alice.MetricUpdate{{Name: "metric.name", CurrentReading: 20}}
	recommendation, _ = thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, alice.SCALEUP)

	setupThresholdStrategyTest()
	config.Set("thresholds.metric.name.foo", "invalid")
	mockResponse = []alice.MetricUpdate{{Name: "metric.name", CurrentReading: 0}}
	_, err := thresholdStrategy.Evaluate()
	assert.Error(t, err)

	config.Set("thresholds.metric.name.min", 5)
	recommendation, _ = thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, alice.SCALEDOWN)

	mockResponse = []alice.MetricUpdate{{Name: "metric.name", CurrentReading: 10}}
	recommendation, _ = thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, alice.HOLD)

	config.Set("thresholds.metric.name.max", 6)
	recommendation, _ = thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, alice.SCALEUP)
}

func TestThresholdStrategy_EvaluateSingleMetricInverted(t *testing.T) {
	setupThresholdStrategyTest()

	config.Set("thresholds.metric.name.min", 5)
	config.Set("thresholds.metric.name.max", 15)
	config.Set("thresholds.metric.name.invert_scaling", true)

	mockResponse = []alice.MetricUpdate{{Name: "metric.name", CurrentReading: 0}}
	recommendation, _ := thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, alice.SCALEUP)

	mockResponse = []alice.MetricUpdate{{Name: "metric.name", CurrentReading: 10}}
	recommendation, _ = thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, alice.HOLD)

	mockResponse = []alice.MetricUpdate{{Name: "metric.name", CurrentReading: 20}}
	recommendation, _ = thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, alice.SCALEDOWN)
}

func TestThresholdStrategy_EvaluateMultipleMetrics(t *testing.T) {
	setupThresholdStrategyTest()

	config.Set("thresholds.metric.one.min", 5)
	config.Set("thresholds.metric.one.max", 15)
	config.Set("thresholds.metric.two.min", 5)
	config.Set("thresholds.metric.two.max", 15)

	mockResponse = []alice.MetricUpdate{
		{Name: "metric.one", CurrentReading: 0},
		{Name: "metric.two", CurrentReading: 0},
	}
	recommendation, _ := thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, alice.SCALEDOWN)

	mockResponse = []alice.MetricUpdate{
		{Name: "metric.one", CurrentReading: 0},
		{Name: "metric.two", CurrentReading: 10},
	}
	recommendation, _ = thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, alice.HOLD)

	mockResponse = []alice.MetricUpdate{
		{Name: "metric.one", CurrentReading: 0},
		{Name: "metric.two", CurrentReading: 20},
	}
	recommendation, _ = thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, alice.SCALEUP)

	mockResponse = []alice.MetricUpdate{
		{Name: "metric.one", CurrentReading: 10},
		{Name: "metric.two", CurrentReading: 0},
	}
	recommendation, _ = thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, alice.HOLD)

	mockResponse = []alice.MetricUpdate{
		{Name: "metric.one", CurrentReading: 20},
		{Name: "metric.two", CurrentReading: 0},
	}
	recommendation, _ = thresholdStrategy.Evaluate()
	assert.Equal(t, *recommendation, alice.SCALEUP)

}
