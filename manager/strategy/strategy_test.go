package strategy_test

import (
	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/autoscaler/manager/inventory"
	"github.com/notonthehighstreet/autoscaler/manager/monitor"
	"github.com/notonthehighstreet/autoscaler/manager/strategy"
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
	strategy.Register("mock", strategy.MockNew)
	config.Set("name", "mock")
}

func TestNew(t *testing.T) {
	setupTest()
	m, _ := monitor.MockNew(config, log)
	i, _ := inventory.MockNew(config, log)
	s, _ := strategy.New(config, i, m, log)
	assert.IsType(t, &strategy.MockStrategy{}, s)
}
