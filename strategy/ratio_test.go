package strategy_test

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/autoscaler/inventory"
	"github.com/notonthehighstreet/autoscaler/monitor"
	"github.com/notonthehighstreet/autoscaler/strategy"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

var metricUpdates []monitor.MetricUpdate
var ratioStrategy *strategy.Ratio

func setupRatioTest() {
	log = logrus.WithFields(logrus.Fields{
		"manager":  "Mock",
		"strategy": "RatioStrategy",
	})
	config = viper.New()
	metricUpdates = []monitor.MetricUpdate{}
	//metricUpdates = append(metricUpdates, monitor.MetricUpdate{Name: "connections", CurrentReading: 200})

	config = viper.New()
	m, _ := monitor.MockNew(config, log)
	mockMonitor = m.(*monitor.MockMonitor)
	i, _ := inventory.MockNew(config, log)
	mockInventory, _ = i.(*inventory.MockInventory)
	r, _ := strategy.NewRatio(config, mockInventory, mockMonitor, log)
	ratioStrategy = r.(*strategy.Ratio)

	mockMonitor.On("GetUpdatedMetrics").Return(&metricUpdates, nil)
}
func TestRatioStrategy_Evaluate(t *testing.T) {
	setupRatioTest()

	mockInventory.On("Total").Return(1, errors.New("foo")).Once()
	recommendation, error := ratioStrategy.Evaluate()
	assert.Error(t, error)

	config.Set("ratios.active_users.metric", 100)
	config.Set("ratios.active_users.inventory", 1)
	metricUpdates = append(metricUpdates, monitor.MetricUpdate{Name: "active_users", CurrentReading: 100})
	mockInventory.On("Total").Return(1, nil).Once()
	recommendation, error = ratioStrategy.Evaluate()
	assert.Nil(t, error)
	assert.Equal(t, *recommendation, strategy.HOLD)
	mockInventory.On("Total").Return(2, nil).Once()
	recommendation, _ = ratioStrategy.Evaluate()
	assert.Equal(t, *recommendation, strategy.SCALEDOWN)
	mockInventory.On("Total").Return(0, nil).Once()
	recommendation, _ = ratioStrategy.Evaluate()
	assert.Equal(t, *recommendation, strategy.SCALEUP)

	config.Set("ratios.active_users.metric", 110)
	config.Set("ratios.active_users.inventory", 100)
	metricUpdates[0] = monitor.MetricUpdate{Name: "active_users", CurrentReading: 220}
	mockInventory.On("Total").Return(200, nil).Once()
	recommendation, error = ratioStrategy.Evaluate()
	assert.Nil(t, error)
	assert.Equal(t, *recommendation, strategy.HOLD)
	mockInventory.On("Total").Return(201, nil).Once()
	recommendation, _ = ratioStrategy.Evaluate()
	assert.Equal(t, *recommendation, strategy.SCALEDOWN)
	mockInventory.On("Total").Return(199, nil).Once()
	recommendation, _ = ratioStrategy.Evaluate()
	assert.Equal(t, *recommendation, strategy.SCALEUP)

	config.Set("ratios.connections.metric", 120)
	config.Set("ratios.connections.inventory", 100)
	metricUpdates = append(metricUpdates, monitor.MetricUpdate{Name: "connections", CurrentReading: 220})
	mockInventory.On("Total").Return(200, nil).Once()
	recommendation, error = ratioStrategy.Evaluate()
	assert.Nil(t, error)
	assert.Equal(t, *recommendation, strategy.HOLD)
	mockInventory.On("Total").Return(201, nil).Once()
	recommendation, _ = ratioStrategy.Evaluate()
	assert.Equal(t, *recommendation, strategy.SCALEDOWN)
	mockInventory.On("Total").Return(199, nil).Once()
	recommendation, _ = ratioStrategy.Evaluate()
	assert.Equal(t, *recommendation, strategy.SCALEUP)
}