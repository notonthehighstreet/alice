package marathon_test

import (
	"github.com/Sirupsen/logrus"
	marathonclient "github.com/gambol99/go-marathon"
	"github.com/notonthehighstreet/autoscaler/manager/inventory/marathon"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"

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
	config.Set("url", "http://foo.com:8080")
	i, _ := marathon.New(config, log)
	inv = i.(*marathon.MarathonInventory)
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
	assert.NoError(t, inv.Increase())

	inv.Config.Set("maximum_instances", 1)
	assert.Error(t, inv.Increase())
}

func TestMarathon_Decrease(t *testing.T) {
	setupTest()
	deployment := marathonclient.DeploymentID{}
	mockClient.On("ScaleApplicationInstances").Return(deployment, nil)
	assert.NoError(t, inv.Decrease())

	inv.Config.Set("minimum_instances", 1)
	assert.Error(t, inv.Decrease())
}

func TestMarathon_Status(t *testing.T) {
	setupTest()
	s, _ := inv.Status()
	assert.Equal(t, inventory.OK, s)
}

func TestSettleDownTime(t *testing.T) {
	setupTest()
	deployment := marathonclient.DeploymentID{}
	mockClient.On("ScaleApplicationInstances").Return(deployment, nil)
	inv.Config.Set("settle_down_period", "5m")
	assert.Nil(t, inv.Increase())
	s, _ := inv.Status()
	assert.Equal(t, inventory.UPDATING, s)
	assert.Error(t, inv.Decrease())
	assert.Error(t, inv.Increase())
}
