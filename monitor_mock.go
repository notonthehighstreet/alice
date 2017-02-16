package autoscaler

import (
	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/mock"
)

type MockMonitor struct {
	mock.Mock
}

func (m *MockMonitor) GetUpdatedMetrics(names []string) (*[]MetricUpdate, error) {
	args := m.Mock.Called()
	return args.Get(0).(*[]MetricUpdate), args.Error(1)
}

func NewMockMonitor(_ *viper.Viper, _ *logrus.Entry) (Monitor, error) {
	return &MockMonitor{}, nil
}
