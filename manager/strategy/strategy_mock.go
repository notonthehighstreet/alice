package strategy

import (
	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/autoscaler/manager/inventory"
	"github.com/notonthehighstreet/autoscaler/manager/monitor"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/mock"
)

type MockStrategy struct {
	mock.Mock
}

func (m *MockStrategy) Evaluate() (*Recommendation, error) {
	args := m.Mock.Called()
	return args.Get(0).(*Recommendation), args.Error(1)
}

func MockNew(_ *viper.Viper, _ inventory.Inventory, _ monitor.Monitor, _ *logrus.Entry) (Strategy, error) {
	return &MockStrategy{}, nil
}
