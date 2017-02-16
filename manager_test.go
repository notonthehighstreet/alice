package autoscaler_test

import (
	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/autoscaler"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

var config = viper.New()
var log = logrus.WithFields(logrus.Fields{
	"manager": "Test",
})
var inv autoscaler.MockInventory
var mon autoscaler.MockMonitor
var str autoscaler.MockStrategy
var recommendation autoscaler.Recommendation
var man autoscaler.Manager

func init() {
	// Register plugins at load time
	autoscaler.RegisterInventory("mock", autoscaler.NewMockInventory)
	autoscaler.RegisterMonitor("mock", autoscaler.NewMockMonitor)
	autoscaler.RegisterStrategy("mock", autoscaler.NewMockStrategy)
}

func setupManagerTest() {
	config.Set("monitor.name", "mock")
	config.Set("inventory.name", "mock")
	config.Set("strategy.name", "mock")
	recommendation = autoscaler.HOLD
	man = autoscaler.Manager{Strategy: &str, Inventory: &inv, Logger: log, Config: config}

}

func TestManager_Run(t *testing.T) {
	setupManagerTest()
	str.On("Evaluate").Return(&recommendation, nil).Once()
	assert.NoError(t, man.Run())
}

func TestScaleUpDisabled(t *testing.T) {
	setupManagerTest()
	config.Set("scale_up", false)
	config.Set("scale_down", true)
	recommendation = autoscaler.SCALEUP
	str.On("Evaluate").Return(&recommendation, nil).Once()
	inv.On("Increase").Return(nil).Once()
	man.Run()
	inv.AssertNotCalled(t, "Increase")
}

func TestScaleDownDisabled(t *testing.T) {
	setupManagerTest()
	config.Set("scale_up", true)
	config.Set("scale_down", false)
	recommendation = autoscaler.SCALEDOWN
	str.On("Evaluate").Return(&recommendation, nil).Once()
	inv.On("Decrease").Return(nil).Once()
	man.Run()
	inv.AssertNotCalled(t, "Decrease")
}
