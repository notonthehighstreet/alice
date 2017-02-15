package inventory_test

import (
	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/autoscaler/inventory"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

var config = viper.New()
var log *logrus.Entry

func setupTest() {
	log = logrus.WithFields(logrus.Fields{
		"manager": "Mock",
	})
	inventory.Register("mock", inventory.MockNew)
	config.Set("name", "mock")
}

func TestNew(t *testing.T) {
	setupTest()
	i, _ := inventory.New(config, log)
	assert.IsType(t, &inventory.MockInventory{}, i)
}
