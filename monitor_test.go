package alice_test

import (
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/alice"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockMonitor struct {
	mock.Mock
}

func (m *MockMonitor) GetUpdatedMetrics(names []string) (*[]alice.MetricUpdate, error) {
	args := m.Mock.Called()
	return args.Get(0).(*[]alice.MetricUpdate), args.Error(1)
}

func NewMockMonitor(_ *viper.Viper, _ *logrus.Entry) (alice.Monitor, error) {
	return &MockMonitor{}, nil
}

func setupMonitorTest() {
	log = logrus.WithFields(logrus.Fields{
		"manager": "Mock",
	})
	alice.RegisterMonitor("mock", NewMockMonitor)
	config.Set("name", "mock")
}

func TestNewMonitor(t *testing.T) {
	setupMonitorTest()
	i, _ := alice.NewMonitor(config, log)
	assert.IsType(t, &MockMonitor{}, i)
}
