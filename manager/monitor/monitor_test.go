package monitor_test

import (
	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/autoscaler/manager/monitor"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

var config = viper.New()
var log = logrus.Entry{}

func setupTest() {
	monitor.Register("mock", monitor.MockNew)
	config.Set("name", "mock")
}

func TestNew(t *testing.T) {
	setupTest()
	i := monitor.New(config, &log)
	assert.IsType(t, &monitor.MockMonitor{}, i)
}