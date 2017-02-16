package alice_test

import (
	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/alice"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type MockInventory struct {
	mock.Mock
}

func (m *MockInventory) Total() (int, error) {
	args := m.Mock.Called()
	return args.Get(0).(int), args.Error(1)
}
func (m *MockInventory) Increase() error {
	args := m.Mock.Called()
	return args.Error(0)
}
func (m *MockInventory) Decrease() error {
	args := m.Mock.Called()
	return args.Error(0)
}
func (m *MockInventory) Status() (alice.Status, error) {
	args := m.Mock.Called()
	return args.Get(0).(alice.Status), nil
}

func NewMockInventory(_ *viper.Viper, _ *logrus.Entry) (alice.Inventory, error) {
	return &MockInventory{}, nil
}

func setupInventoryTest() {
	log = logrus.WithFields(logrus.Fields{
		"manager": "Mock",
	})
	alice.RegisterInventory("mock", NewMockInventory)
	config.Set("name", "mock")
}

func TestNewInventory(t *testing.T) {
	setupInventoryTest()
	i, _ := alice.NewInventory(config, log)
	assert.IsType(t, &MockInventory{}, i)
}
