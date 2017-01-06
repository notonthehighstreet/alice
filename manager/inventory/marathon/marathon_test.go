package marathon_test

import (
	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
	"github.com/notonthehighstreet/autoscaler/manager/inventory/marathon"
	marathonclient "github.com/gambol99/go-marathon"

	"github.com/notonthehighstreet/autoscaler/manager/inventory"
)

var log = logrus.WithFields(logrus.Fields{
	"manager":   "Mock",
	"inventory": "MarathonInventory",
})
var inv *marathon.MarathonInventory
var mockClient marathon.MockMarathonClient

func setupTest() {
	log.Logger.Level = logrus.DebugLevel
	config := viper.New()
	config.Set("app", "notonthehighstreet-admin")
	inv = marathon.New(config, log).(*marathon.MarathonInventory)
	inv.Client = &mockClient
	instances := 1
	app := marathonclient.Application{Instances: &instances}
	mockClient.On("ApplicationBy").Return(app, nil)
}

func TestMarathon_Total(t *testing.T) {
	setupTest()
	total, _ := inv.Total()
	assert.Equal(t, total, 1)
}

func TestMarathonInventory_Scale(t *testing.T) {
	setupTest()
	deployment := marathonclient.DeploymentID{}
	mockClient.On("ScaleApplicationInstances").Return(deployment, nil)
	err := inv.Scale(+1)
	assert.NoError(t, err)
}

func TestMarathon_Increase(t *testing.T) {
	setupTest()
	deployment := marathonclient.DeploymentID{}
	mockClient.On("ScaleApplicationInstances").Return(deployment, nil)
	assert.Nil(t, inv.Increase())
}

func TestMarathon_Decrease(t *testing.T) {
	setupTest()
	deployment := marathonclient.DeploymentID{}
	mockClient.On("ScaleApplicationInstances").Return(deployment, nil)
	assert.Nil(t, inv.Decrease())
}

func TestMarathon_Status(t *testing.T) {
	setupTest()
	assert.Equal(t, inventory.OK, inv.Status())
}

func TestSettleDownTime(t *testing.T) {
	setupTest()
	deployment := marathonclient.DeploymentID{}
	mockClient.On("ScaleApplicationInstances").Return(deployment, nil)
	inv.Config.Set("settle_down_period", "5m")
	assert.Nil(t, inv.Increase())
	assert.Equal(t, inventory.UPDATING, inv.Status())
	assert.Error(t, inv.Decrease())
	assert.Error(t, inv.Increase())
}
