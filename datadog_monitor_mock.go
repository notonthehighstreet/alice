package autoscaler

import (
	"github.com/stretchr/testify/mock"
	"github.com/zorkian/go-datadog-api"
)

type MockDatadogClient struct {
	mock.Mock
}

func (d *MockDatadogClient) Validate() (bool, error) {
	args := d.Mock.Called()
	return args.Get(0).(bool), args.Error(1)
}

func (d *MockDatadogClient) QueryMetrics(from, to int64, query string) ([]datadog.Series, error) {
	args := d.Mock.Called()
	return args.Get(0).([]datadog.Series), args.Error(1)
}
