package inventory

import "github.com/stretchr/testify/mock"

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
func (m *MockInventory) Status() Status {
	args := m.Mock.Called()
	return args.Get(0).(Status)
}