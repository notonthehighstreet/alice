package autoscaler

import (
	"github.com/Sirupsen/logrus"
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

func NewMockStrategy(_ *viper.Viper, _ Inventory, _ Monitor, _ *logrus.Entry) (Strategy, error) {
	return &MockStrategy{}, nil
}
