package alice_test

import (
	"errors"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/alice"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var metricUpdates []alice.MetricUpdate
var ratioStrategy *alice.RatioStrategy

func setupRatioStrategyTest() {
	log = logrus.WithFields(logrus.Fields{
		"manager":  "Mock",
		"strategy": "RatioStrategy",
	})
	config = viper.New()
	metricUpdates = []alice.MetricUpdate{}

	config = viper.New()
	m, _ := NewMockMonitor(config, log)
	mockMonitor = m.(*MockMonitor)
	i, _ := NewMockInventory(config, log)
	mockInventory, _ = i.(*MockInventory)
	r, _ := alice.NewRatioStrategy(config, mockInventory, mockMonitor, log)
	ratioStrategy = r.(*alice.RatioStrategy)

	mockMonitor.On("GetUpdatedMetrics").Return(&metricUpdates, nil)
}
func TestRatioStrategy_Evaluate(t *testing.T) {
	setupRatioStrategyTest()

	mockInventory.On("Total").Return(1, errors.New("foo")).Once()
	recommendation, error := ratioStrategy.Evaluate()
	assert.Error(t, error)

	config.Set("ratios.active_users.metric", 100)
	config.Set("ratios.active_users.inventory", 1)
	metricUpdates = append(metricUpdates, alice.MetricUpdate{Name: "active_users", CurrentReading: 100})
	mockInventory.On("Total").Return(1, nil).Once()
	recommendation, error = ratioStrategy.Evaluate()
	assert.Nil(t, error)
	assert.Equal(t, *recommendation, alice.HOLD)
	mockInventory.On("Total").Return(2, nil).Once()
	recommendation, _ = ratioStrategy.Evaluate()
	assert.Equal(t, *recommendation, alice.SCALEDOWN)
	mockInventory.On("Total").Return(0, nil).Once()
	recommendation, _ = ratioStrategy.Evaluate()
	assert.Equal(t, *recommendation, alice.SCALEUP)

	config.Set("ratios.active_users.metric", 110)
	config.Set("ratios.active_users.inventory", 100)
	metricUpdates[0] = alice.MetricUpdate{Name: "active_users", CurrentReading: 220}
	mockInventory.On("Total").Return(200, nil).Once()
	recommendation, error = ratioStrategy.Evaluate()
	assert.Nil(t, error)
	assert.Equal(t, *recommendation, alice.HOLD)
	mockInventory.On("Total").Return(201, nil).Once()
	recommendation, _ = ratioStrategy.Evaluate()
	assert.Equal(t, *recommendation, alice.SCALEDOWN)
	mockInventory.On("Total").Return(199, nil).Once()
	recommendation, _ = ratioStrategy.Evaluate()
	assert.Equal(t, *recommendation, alice.SCALEUP)

	config.Set("ratios.connections.metric", 120)
	config.Set("ratios.connections.inventory", 100)
	metricUpdates = append(metricUpdates, alice.MetricUpdate{Name: "connections", CurrentReading: 220})
	mockInventory.On("Total").Return(200, nil).Once()
	recommendation, error = ratioStrategy.Evaluate()
	assert.Nil(t, error)
	assert.Equal(t, *recommendation, alice.HOLD)
	mockInventory.On("Total").Return(201, nil).Once()
	recommendation, _ = ratioStrategy.Evaluate()
	assert.Equal(t, *recommendation, alice.SCALEDOWN)
	mockInventory.On("Total").Return(199, nil).Once()
	recommendation, _ = ratioStrategy.Evaluate()
	assert.Equal(t, *recommendation, alice.SCALEUP)
}
