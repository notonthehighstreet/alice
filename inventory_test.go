package alice_test

import (
	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/alice"
	"github.com/stretchr/testify/assert"
	"testing"
)

func setupInventoryTest() {
	log = logrus.WithFields(logrus.Fields{
		"manager": "Mock",
	})
	alice.RegisterInventory("mock", alice.NewMockInventory)
	config.Set("name", "mock")
}

func TestNewInventory(t *testing.T) {
	setupInventoryTest()
	i, _ := alice.NewInventory(config, log)
	assert.IsType(t, &alice.MockInventory{}, i)
}
