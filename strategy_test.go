package alice_test

import (
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/alice"
	"github.com/stretchr/testify/assert"
)

var mockInventory *alice.MockInventory
var mockMonitor *alice.MockMonitor

func setupStrategyTest() {
	log = logrus.WithFields(logrus.Fields{
		"manager": "Mock",
	})
	alice.RegisterStrategy("mock", alice.NewMockStrategy)
	config.Set("name", "mock")
	m, _ := alice.NewMockMonitor(config, log)
	mockMonitor = m.(*alice.MockMonitor)
	i, _ := alice.NewMockInventory(config, log)
	mockInventory, _ = i.(*alice.MockInventory)
}

func TestNewStrategy(t *testing.T) {
	setupStrategyTest()
	s, _ := alice.NewStrategy(config, mockInventory, mockMonitor, log)
	assert.IsType(t, &alice.MockStrategy{}, s)
}
