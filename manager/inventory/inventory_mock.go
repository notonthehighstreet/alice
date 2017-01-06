package inventory

import (
	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/mock"
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
func (m *MockInventory) Status() (Status, error) {
	args := m.Mock.Called()
	return args.Get(0).(Status), nil
}

func MockNew(_ *viper.Viper, _ *logrus.Entry) Inventory {
	return &MockInventory{}
}
