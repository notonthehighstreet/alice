package autoscaler_test

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/autoscaler"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

var metricUpdates []autoscaler.MetricUpdate
var ratioStrategy *autoscaler.RatioStrategy

func setupRatioStrategyTest() {
	log = logrus.WithFields(logrus.Fields{
		"manager":  "Mock",
		"strategy": "RatioStrategy",
	})
	config = viper.New()
	metricUpdates = []autoscaler.MetricUpdate{}

	config = viper.New()
	m, _ := autoscaler.NewMockMonitor(config, log)
	mockMonitor = m.(*autoscaler.MockMonitor)
	i, _ := autoscaler.NewMockInventory(config, log)
	mockInventory, _ = i.(*autoscaler.MockInventory)
	r, _ := autoscaler.NewRatioStrategy(config, mockInventory, mockMonitor, log)
	ratioStrategy = r.(*autoscaler.RatioStrategy)

	mockMonitor.On("GetUpdatedMetrics").Return(&metricUpdates, nil)
}
func TestRatioStrategy_Evaluate(t *testing.T) {
	setupRatioStrategyTest()

	mockInventory.On("Total").Return(1, errors.New("foo")).Once()
	recommendation, error := ratioStrategy.Evaluate()
	assert.Error(t, error)

	config.Set("ratios.active_users.metric", 100)
	config.Set("ratios.active_users.inventory", 1)
	metricUpdates = append(metricUpdates, autoscaler.MetricUpdate{Name: "active_users", CurrentReading: 100})
	mockInventory.On("Total").Return(1, nil).Once()
	recommendation, error = ratioStrategy.Evaluate()
	assert.Nil(t, error)
	assert.Equal(t, *recommendation, autoscaler.HOLD)
	mockInventory.On("Total").Return(2, nil).Once()
	recommendation, _ = ratioStrategy.Evaluate()
	assert.Equal(t, *recommendation, autoscaler.SCALEDOWN)
	mockInventory.On("Total").Return(0, nil).Once()
	recommendation, _ = ratioStrategy.Evaluate()
	assert.Equal(t, *recommendation, autoscaler.SCALEUP)

	config.Set("ratios.active_users.metric", 110)
	config.Set("ratios.active_users.inventory", 100)
	metricUpdates[0] = autoscaler.MetricUpdate{Name: "active_users", CurrentReading: 220}
	mockInventory.On("Total").Return(200, nil).Once()
	recommendation, error = ratioStrategy.Evaluate()
	assert.Nil(t, error)
	assert.Equal(t, *recommendation, autoscaler.HOLD)
	mockInventory.On("Total").Return(201, nil).Once()
	recommendation, _ = ratioStrategy.Evaluate()
	assert.Equal(t, *recommendation, autoscaler.SCALEDOWN)
	mockInventory.On("Total").Return(199, nil).Once()
	recommendation, _ = ratioStrategy.Evaluate()
	assert.Equal(t, *recommendation, autoscaler.SCALEUP)

	config.Set("ratios.connections.metric", 120)
	config.Set("ratios.connections.inventory", 100)
	metricUpdates = append(metricUpdates, autoscaler.MetricUpdate{Name: "connections", CurrentReading: 220})
	mockInventory.On("Total").Return(200, nil).Once()
	recommendation, error = ratioStrategy.Evaluate()
	assert.Nil(t, error)
	assert.Equal(t, *recommendation, autoscaler.HOLD)
	mockInventory.On("Total").Return(201, nil).Once()
	recommendation, _ = ratioStrategy.Evaluate()
	assert.Equal(t, *recommendation, autoscaler.SCALEDOWN)
	mockInventory.On("Total").Return(199, nil).Once()
	recommendation, _ = ratioStrategy.Evaluate()
	assert.Equal(t, *recommendation, autoscaler.SCALEUP)
}
