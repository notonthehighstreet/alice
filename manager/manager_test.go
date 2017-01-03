package manager_test

import (
	"github.com/notonthehighstreet/autoscaler/manager/inventory"
	"github.com/notonthehighstreet/autoscaler/manager/monitor"
	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/autoscaler/manager/strategy"
	"testing"
	"github.com/spf13/viper"
	"github.com/notonthehighstreet/autoscaler/manager"
	"github.com/stretchr/testify/assert"
)

var conf = viper.New()
var log = logrus.WithFields(logrus.Fields{
	"manager":  "Test",
})
var inv inventory.MockInventory
var mon monitor.MockMonitor
var str strategy.MockStrategy
var recommendation strategy.Recommendation
var man manager.Manager

func init() {
	// Register plugins at load time
	inventory.Register("mock", inventory.MockNew)
	monitor.Register("mock", monitor.MockNew)
	strategy.Register("mock", strategy.MockNew)
}

func setupTest()  {
	conf.Set("monitor.name", "mock")
	conf.Set("inventory.name", "mock")
	conf.Set("strategy.name", "mock")
	recommendation = strategy.HOLD
	man = manager.Manager{Strategy: &str, Inventory: &inv, Logger: log, Config: conf}

}


func TestManager_Run(t *testing.T) {
	setupTest()
	str.On("Evaluate").Return(&recommendation, nil).Once()
	assert.NoError(t, man.Run())
}

func TestScaleUpDisabled(t *testing.T) {
	setupTest()
	conf.Set("scale_up", false)
	conf.Set("scale_down", true)
	recommendation = strategy.SCALEUP
	str.On("Evaluate").Return(&recommendation, nil).Once()
	inv.On("Increase").Return(nil).Once()
	man.Run()
	inv.AssertNotCalled(t, "Increase")
}

func TestScaleDownDisabled(t *testing.T) {
	setupTest()
	conf.Set("scale_up", true)
	conf.Set("scale_down", false)
	recommendation = strategy.SCALEDOWN
	str.On("Evaluate").Return(&recommendation, nil).Once()
	inv.On("Decrease").Return(nil).Once()
	man.Run()
	inv.AssertNotCalled(t, "Decrease")
}