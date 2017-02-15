package monitor_test

import (
	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/autoscaler/monitor"
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
	monitor.Register("mock", monitor.MockNew)
	config.Set("name", "mock")
}

func TestNew(t *testing.T) {
	setupTest()
	i, _ := monitor.New(config, log)
	assert.IsType(t, &monitor.MockMonitor{}, i)
}
