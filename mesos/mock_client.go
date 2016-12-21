package mesos

import (
	"github.com/stretchr/testify/mock"
	"github.com/andygrunwald/megos"
)

type MockMesosClient struct {
	mock.Mock
}

func (m *MockMesosClient) GetStateFromLeader() (*megos.State, error) {
	args := m.Mock.Called()
	state := args.Get(0).(megos.State)
	return &state, args.Error(1)
}
