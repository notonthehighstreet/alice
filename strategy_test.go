package autoscaler_test

import (
	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/autoscaler"
	"github.com/stretchr/testify/assert"
	"testing"
)

var mockInventory *autoscaler.MockInventory
var mockMonitor *autoscaler.MockMonitor

func setupStrategyTest() {
	log = logrus.WithFields(logrus.Fields{
		"manager": "Mock",
	})
	autoscaler.RegisterStrategy("mock", autoscaler.NewMockStrategy)
	config.Set("name", "mock")
	m, _ := autoscaler.NewMockMonitor(config, log)
	mockMonitor = m.(*autoscaler.MockMonitor)
	i, _ := autoscaler.NewMockInventory(config, log)
	mockInventory, _ = i.(*autoscaler.MockInventory)
}

func TestNewStrategy(t *testing.T) {
	setupStrategyTest()
	s, _ := autoscaler.NewStrategy(config, mockInventory, mockMonitor, log)
	assert.IsType(t, &autoscaler.MockStrategy{}, s)
}
