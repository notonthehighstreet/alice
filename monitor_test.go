package autoscaler_test

import (
	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/autoscaler"
	"github.com/stretchr/testify/assert"
	"testing"
)

func setupMonitorTest() {
	log = logrus.WithFields(logrus.Fields{
		"manager": "Mock",
	})
	autoscaler.RegisterMonitor("mock", autoscaler.NewMockMonitor)
	config.Set("name", "mock")
}

func TestNewMonitor(t *testing.T) {
	setupMonitorTest()
	i, _ := autoscaler.NewMonitor(config, log)
	assert.IsType(t, &autoscaler.MockMonitor{}, i)
}
