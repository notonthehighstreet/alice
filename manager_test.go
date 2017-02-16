package alice_test

import (
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/alice"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var config = viper.New()
var log = logrus.WithFields(logrus.Fields{
	"manager": "Test",
})
var inv alice.MockInventory
var mon alice.MockMonitor
var str alice.MockStrategy
var recommendation alice.Recommendation
var man alice.Manager

func init() {
	// Register plugins at load time
	alice.RegisterInventory("mock", alice.NewMockInventory)
	alice.RegisterMonitor("mock", alice.NewMockMonitor)
	alice.RegisterStrategy("mock", alice.NewMockStrategy)
}

func setupManagerTest() {
	config.Set("monitor.name", "mock")
	config.Set("inventory.name", "mock")
	config.Set("strategy.name", "mock")
	recommendation = alice.HOLD
	man = alice.Manager{Strategy: &str, Inventory: &inv, Logger: log, Config: config}

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
	recommendation = alice.SCALEUP
	str.On("Evaluate").Return(&recommendation, nil).Once()
	inv.On("Increase").Return(nil).Once()
	man.Run()
	inv.AssertNotCalled(t, "Increase")
}

func TestScaleDownDisabled(t *testing.T) {
	setupManagerTest()
	config.Set("scale_up", true)
	config.Set("scale_down", false)
	recommendation = alice.SCALEDOWN
	str.On("Evaluate").Return(&recommendation, nil).Once()
	inv.On("Decrease").Return(nil).Once()
	man.Run()
	inv.AssertNotCalled(t, "Decrease")
}
