package alice_test

import (
	"github.com/Sirupsen/logrus"
	"github.com/gambol99/go-marathon"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"

	"github.com/notonthehighstreet/alice"
)

type MockMarathonClient struct {
	mock.Mock
}

func (m *MockMarathonClient) ApplicationBy(name string, opts *marathon.GetAppOpts) (*marathon.Application, error) {
	args := m.Mock.Called()
	app := args.Get(0).(marathon.Application)
	return &app, args.Error(1)
}

func (m *MockMarathonClient) ScaleApplicationInstances(name string, instances int, force bool) (*marathon.DeploymentID, error) {
	args := m.Mock.Called()
	dep := args.Get(0).(marathon.DeploymentID)
	return &dep, args.Error(1)
}

var marathonInv *alice.MarathonInventory
var mockClient MockMarathonClient

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
	app := marathon.Application{Instances: &instances}
	mockClient.On("ApplicationBy").Return(app, nil)
}

func TestMarathonInventory_Total(t *testing.T) {
	setupMarathonInventoryTest()
	total, _ := marathonInv.Total()
	assert.Equal(t, total, 1)
}

func TestMarathonInventory_Scale(t *testing.T) {
	setupMarathonInventoryTest()
	deployment := marathon.DeploymentID{}
	mockClient.On("ScaleApplicationInstances").Return(deployment, nil)
	err := marathonInv.Scale(+1)
	assert.NoError(t, err)
}

func TestMarathonInventory_Increase(t *testing.T) {
	setupMarathonInventoryTest()
	deployment := marathon.DeploymentID{}
	mockClient.On("ScaleApplicationInstances").Return(deployment, nil)
	assert.NoError(t, marathonInv.Increase())

	marathonInv.Config.Set("maximum_instances", 1)
	assert.Error(t, marathonInv.Increase())
}

func TestMarathonInventory_Decrease(t *testing.T) {
	setupMarathonInventoryTest()
	deployment := marathon.DeploymentID{}
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
	deployment := marathon.DeploymentID{}
	mockClient.On("ScaleApplicationInstances").Return(deployment, nil)
	marathonInv.Config.Set("settle_down_period", "5m")
	assert.Nil(t, marathonInv.Increase())
	s, _ := marathonInv.Status()
	assert.Equal(t, alice.UPDATING, s)
	assert.Error(t, marathonInv.Decrease())
	assert.Error(t, marathonInv.Increase())
}
