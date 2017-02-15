package inventory_test

import (
	"github.com/Sirupsen/logrus"
	marathonclient "github.com/gambol99/go-marathon"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/notonthehighstreet/autoscaler/manager/inventory"
)

var marathonInv *inventory.Marathon
var mockClient inventory.MockMarathonClient

func setupMarathonTest() {
	log = logrus.WithFields(logrus.Fields{
		"manager":   "Mock",
		"inventory": "MarathonInventory",
	})
	log.Logger.Level = logrus.DebugLevel
	config := viper.New()
	config.Set("app", "notonthehighstreet-admin")
	config.Set("url", "http://foo.com:8080")
	i, _ := inventory.NewMarathon(config, log)
	marathonInv = i.(*inventory.Marathon)
	marathonInv.Client = &mockClient
	instances := 1
	app := marathonclient.Application{Instances: &instances}
	mockClient.On("ApplicationBy").Return(app, nil)
}

func TestMarathon_Total(t *testing.T) {
	setupMarathonTest()
	total, _ := marathonInv.Total()
	assert.Equal(t, total, 1)
}

func TestMarathonInventory_Scale(t *testing.T) {
	setupMarathonTest()
	deployment := marathonclient.DeploymentID{}
	mockClient.On("ScaleApplicationInstances").Return(deployment, nil)
	err := marathonInv.Scale(+1)
	assert.NoError(t, err)
}

func TestMarathon_Increase(t *testing.T) {
	setupMarathonTest()
	deployment := marathonclient.DeploymentID{}
	mockClient.On("ScaleApplicationInstances").Return(deployment, nil)
	assert.NoError(t, marathonInv.Increase())

	marathonInv.Config.Set("maximum_instances", 1)
	assert.Error(t, marathonInv.Increase())
}

func TestMarathon_Decrease(t *testing.T) {
	setupMarathonTest()
	deployment := marathonclient.DeploymentID{}
	mockClient.On("ScaleApplicationInstances").Return(deployment, nil)
	assert.NoError(t, marathonInv.Decrease())

	marathonInv.Config.Set("minimum_instances", 1)
	assert.Error(t, marathonInv.Decrease())
}

func TestMarathon_Status(t *testing.T) {
	setupMarathonTest()
	s, _ := marathonInv.Status()
	assert.Equal(t, inventory.OK, s)
}

func TestMarathon_SettleDownTime(t *testing.T) {
	setupMarathonTest()
	deployment := marathonclient.DeploymentID{}
	mockClient.On("ScaleApplicationInstances").Return(deployment, nil)
	marathonInv.Config.Set("settle_down_period", "5m")
	assert.Nil(t, marathonInv.Increase())
	s, _ := marathonInv.Status()
	assert.Equal(t, inventory.UPDATING, s)
	assert.Error(t, marathonInv.Decrease())
	assert.Error(t, marathonInv.Increase())
}
