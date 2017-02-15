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

var mockInventory *inventory.MockInventory
var mockMonitor *monitor.MockMonitor
var config = viper.New()
var log *logrus.Entry

func setupTest() {
	log = logrus.WithFields(logrus.Fields{
		"manager": "Mock",
	})
	strategy.Register("mock", strategy.MockNew)
	config.Set("name", "mock")
	m, _ := monitor.MockNew(config, log)
	mockMonitor = m.(*monitor.MockMonitor)
	i, _ := inventory.MockNew(config, log)
	mockInventory, _ = i.(*inventory.MockInventory)
}

func TestNew(t *testing.T) {
	setupTest()
	s, _ := strategy.New(config, mockInventory, mockMonitor, log)
	assert.IsType(t, &strategy.MockStrategy{}, s)
}
