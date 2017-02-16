package autoscaler

import (
	"github.com/andygrunwald/megos"
	"github.com/stretchr/testify/mock"
)

type MockMesosClient struct {
	mock.Mock
}

func (m *MockMesosClient) GetStateFromLeader() (*megos.State, error) {
	args := m.Mock.Called()
	state := args.Get(0).(megos.State)
	return &state, args.Error(1)
}

func (m *MockMesosClient) DetermineLeader() (*megos.Pid, error) {
	args := m.Mock.Called()
	state := args.Get(0).(megos.Pid)
	return &state, args.Error(1)
}
