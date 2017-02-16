package autoscaler_test

import (
	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/autoscaler"
	"github.com/stretchr/testify/assert"
	"testing"
)

func setupInventoryTest() {
	log = logrus.WithFields(logrus.Fields{
		"manager": "Mock",
	})
	autoscaler.RegisterInventory("mock", autoscaler.NewMockInventory)
	config.Set("name", "mock")
}

func TestNewInventory(t *testing.T) {
	setupInventoryTest()
	i, _ := autoscaler.NewInventory(config, log)
	assert.IsType(t, &autoscaler.MockInventory{}, i)
}
