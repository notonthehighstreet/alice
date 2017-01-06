package ratio_test

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/autoscaler/manager/inventory"
	"github.com/notonthehighstreet/autoscaler/manager/monitor"
	"github.com/notonthehighstreet/autoscaler/manager/strategy"
	"github.com/notonthehighstreet/autoscaler/manager/strategy/ratio"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

var config = viper.New()
var log = logrus.WithFields(logrus.Fields{
	"manager":  "Mock",
	"strategy": "RatioStrategy",
})
var mockInventory inventory.MockInventory
var mockMonitor monitor.MockMonitor
var metricUpdates []monitor.MetricUpdate
var rat *ratio.RatioStrategy

func setupTest() {
	metricUpdates = []monitor.MetricUpdate{}
	//metricUpdates = append(metricUpdates, monitor.MetricUpdate{Name: "connections", CurrentReading: 200})
	mockMonitor.On("GetUpdatedMetrics").Return(&metricUpdates, nil)

	config = viper.New()
	r, _ := ratio.New(config, &mockInventory, &mockMonitor, log)
	rat = r.(*ratio.RatioStrategy)
}
func TestRatioStrategy_Evaluate(t *testing.T) {
	setupTest()

	mockInventory.On("Total").Return(1, errors.New("foo")).Once()
	recommendation, error := rat.Evaluate()
	assert.Error(t, error)

	config.Set("ratios.active_users.metric", 100)
	config.Set("ratios.active_users.inventory", 1)
	metricUpdates = append(metricUpdates, monitor.MetricUpdate{Name: "active_users", CurrentReading: 100})
	mockInventory.On("Total").Return(1, nil).Once()
	recommendation, error = rat.Evaluate()
	assert.Nil(t, error)
	assert.Equal(t, *recommendation, strategy.HOLD)
	mockInventory.On("Total").Return(2, nil).Once()
	recommendation, _ = rat.Evaluate()
	assert.Equal(t, *recommendation, strategy.SCALEDOWN)
	mockInventory.On("Total").Return(0, nil).Once()
	recommendation, _ = rat.Evaluate()
	assert.Equal(t, *recommendation, strategy.SCALEUP)

	config.Set("ratios.active_users.metric", 110)
	config.Set("ratios.active_users.inventory", 100)
	metricUpdates[0] = monitor.MetricUpdate{Name: "active_users", CurrentReading: 220}
	mockInventory.On("Total").Return(200, nil).Once()
	recommendation, error = rat.Evaluate()
	assert.Nil(t, error)
	assert.Equal(t, *recommendation, strategy.HOLD)
	mockInventory.On("Total").Return(201, nil).Once()
	recommendation, _ = rat.Evaluate()
	assert.Equal(t, *recommendation, strategy.SCALEDOWN)
	mockInventory.On("Total").Return(199, nil).Once()
	recommendation, _ = rat.Evaluate()
	assert.Equal(t, *recommendation, strategy.SCALEUP)

	config.Set("ratios.connections.metric", 120)
	config.Set("ratios.connections.inventory", 100)
	metricUpdates = append(metricUpdates, monitor.MetricUpdate{Name: "connections", CurrentReading: 220})
	mockInventory.On("Total").Return(200, nil).Once()
	recommendation, error = rat.Evaluate()
	assert.Nil(t, error)
	assert.Equal(t, *recommendation, strategy.HOLD)
	mockInventory.On("Total").Return(201, nil).Once()
	recommendation, _ = rat.Evaluate()
	assert.Equal(t, *recommendation, strategy.SCALEDOWN)
	mockInventory.On("Total").Return(199, nil).Once()
	recommendation, _ = rat.Evaluate()
	assert.Equal(t, *recommendation, strategy.SCALEUP)
}
