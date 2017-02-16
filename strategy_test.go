package alice_test

import (
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/alice"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockStrategy struct {
	mock.Mock
}

func (m *MockStrategy) Evaluate() (*alice.Recommendation, error) {
	args := m.Mock.Called()
	return args.Get(0).(*alice.Recommendation), args.Error(1)
}

func NewMockStrategy(_ *viper.Viper, _ alice.Inventory, _ alice.Monitor, _ *logrus.Entry) (alice.Strategy, error) {
	return &MockStrategy{}, nil
}

var mockInventory *MockInventory
var mockMonitor *MockMonitor

func setupStrategyTest() {
	log = logrus.WithFields(logrus.Fields{
		"manager": "Mock",
	})
	alice.RegisterStrategy("mock", NewMockStrategy)
	config.Set("name", "mock")
	m, _ := NewMockMonitor(config, log)
	mockMonitor = m.(*MockMonitor)
	i, _ := NewMockInventory(config, log)
	mockInventory, _ = i.(*MockInventory)
}

func TestNewStrategy(t *testing.T) {
	setupStrategyTest()
	s, _ := alice.NewStrategy(config, mockInventory, mockMonitor, log)
	assert.IsType(t, &MockStrategy{}, s)
}
