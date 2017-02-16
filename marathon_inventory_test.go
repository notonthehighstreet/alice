package alice_test

import (
	"github.com/Sirupsen/logrus"
	marathonclient "github.com/gambol99/go-marathon"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/notonthehighstreet/alice"
)

var marathonInv *alice.MarathonInventory
var mockClient alice.MockMarathonClient

func setupMarathonInventoryTest() {
	log = logrus.WithFields(logrus.Fields{
		"manager":   "Mock",
		"inventory": "MarathonInventory",
	})
	log.Logger.Level = logrus.DebugLevel
	config := viper.New()
	config.Set("app", "notonthehighstreet-admin")
	config.Set("url", "http://foo.com:8080")
	i, _ := alice.NewMarathonInventory(config, log)
	marathonInv = i.(*alice.MarathonInventory)
	marathonInv.Client = &mockClient
	instances := 1
	app := marathonclient.Application{Instances: &instances}
	mockClient.On("ApplicationBy").Return(app, nil)
}

func TestMarathonInventory_Total(t *testing.T) {
	setupMarathonInventoryTest()
	total, _ := marathonInv.Total()
	assert.Equal(t, total, 1)
}

func TestMarathonInventory_Scale(t *testing.T) {
	setupMarathonInventoryTest()
	deployment := marathonclient.DeploymentID{}
	mockClient.On("ScaleApplicationInstances").Return(deployment, nil)
	err := marathonInv.Scale(+1)
	assert.NoError(t, err)
}

func TestMarathonInventory_Increase(t *testing.T) {
	setupMarathonInventoryTest()
	deployment := marathonclient.DeploymentID{}
	mockClient.On("ScaleApplicationInstances").Return(deployment, nil)
	assert.NoError(t, marathonInv.Increase())

	marathonInv.Config.Set("maximum_instances", 1)
	assert.Error(t, marathonInv.Increase())
}

func TestMarathonInventory_Decrease(t *testing.T) {
	setupMarathonInventoryTest()
	deployment := marathonclient.DeploymentID{}
	mockClient.On("ScaleApplicationInstances").Return(deployment, nil)
	assert.NoError(t, marathonInv.Decrease())

	marathonInv.Config.Set("minimum_instances", 1)
	assert.Error(t, marathonInv.Decrease())
}

func TestMarathonInventory_Status(t *testing.T) {
	setupMarathonInventoryTest()
	s, _ := marathonInv.Status()
	assert.Equal(t, alice.OK, s)
}

func TestMarathonInventory_SettleDownTime(t *testing.T) {
	setupMarathonInventoryTest()
	deployment := marathonclient.DeploymentID{}
	mockClient.On("ScaleApplicationInstances").Return(deployment, nil)
	marathonInv.Config.Set("settle_down_period", "5m")
	assert.Nil(t, marathonInv.Increase())
	s, _ := marathonInv.Status()
	assert.Equal(t, alice.UPDATING, s)
	assert.Error(t, marathonInv.Decrease())
	assert.Error(t, marathonInv.Increase())
}
