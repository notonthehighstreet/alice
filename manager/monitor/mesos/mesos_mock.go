package mesos

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
