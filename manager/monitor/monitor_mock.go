package monitor

import "github.com/stretchr/testify/mock"

type MockMonitor struct {
	mock.Mock
}

func (m *MockMonitor) GetUpdatedMetrics(names []string) (*[]MetricUpdate, error) {
	args := m.Mock.Called()
	return args.Get(0).(*[]MetricUpdate), args.Error(1)
}
