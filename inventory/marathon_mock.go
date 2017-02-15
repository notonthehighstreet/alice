package inventory

import (
	"github.com/gambol99/go-marathon"
	"github.com/stretchr/testify/mock"
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
