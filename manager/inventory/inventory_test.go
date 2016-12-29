package inventory_test

import (
	"github.com/notonthehighstreet/autoscaler/manager/inventory"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

var config = viper.New()
var log = logrus.Entry{}

func setupTest() {
	inventory.Register("mock", inventory.MockNew)
	config.Set("name", "mock")
}

func TestNew(t *testing.T) {
	setupTest()
	i := inventory.New(config, &log)
	assert.IsType(t, &inventory.MockInventory{}, i)
}
