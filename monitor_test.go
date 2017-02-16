package alice_test

import (
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/alice"
	"github.com/stretchr/testify/assert"
)

func setupMonitorTest() {
	log = logrus.WithFields(logrus.Fields{
		"manager": "Mock",
	})
	alice.RegisterMonitor("mock", alice.NewMockMonitor)
	config.Set("name", "mock")
}

func TestNewMonitor(t *testing.T) {
	setupMonitorTest()
	i, _ := alice.NewMonitor(config, log)
	assert.IsType(t, &alice.MockMonitor{}, i)
}
