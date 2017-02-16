package alice_test

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/alice"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zorkian/go-datadog-api"
	"testing"
	"time"
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

var datadogMon *alice.DatadogMonitor
var mockDatadogClient MockDatadogClient

func setupDatadogMonitorTest() {
	config = viper.New()
	log = logrus.WithFields(logrus.Fields{
		"manager": "Mock",
		"monitor": "DatadogMonitor",
	})
	config.Set("api_key", "foo")
	config.Set("app_key", "bar")
	config.Set("time_period", "5m")
	log.Logger.Level = logrus.DebugLevel
	mockDatadogClient = MockDatadogClient{}
	m, _ := alice.NewDatadogMonitor(config, log)
	datadogMon = m.(*alice.DatadogMonitor)
	datadogMon.Client = &mockDatadogClient
}

func TestDatadogMonitor_GetUpdatedMetrics(t *testing.T) {
	setupDatadogMonitorTest()
	config.Set("metrics.foo.bar.baz.query", "avg:foo.bar.baz{*}")
	metrics := []string{"foo.bar.baz"}
	mockResponse := []datadog.Series{
		{Points: []datadog.DataPoint{
			{float64(time.Now().Unix() - 2), 0.9},
			{float64(time.Now().Unix() - 1), 0.3},
			{float64(time.Now().Unix()), 0.5},
		}},
	}
	mockDatadogClient.On("Validate").Return(true, nil)
	mockDatadogClient.On("QueryMetrics").Return(mockResponse, nil)
	vp, err := datadogMon.GetUpdatedMetrics(metrics)
	assert.NoError(t, err)
	val := *vp
	assert.Equal(t, 1, len(val))
	assert.Equal(t, 0.5, val[0].CurrentReading)
}

func TestDatadogMonitor_GetUpdatedMetricsNoData(t *testing.T) {
	setupDatadogMonitorTest()
	metrics := []string{"foo.bar.baz"}
	mockDatadogClient.On("Validate").Return(true, nil)

	mockResponse := []datadog.Series{}
	mockDatadogClient.On("QueryMetrics").Return(mockResponse, nil).Once()
	_, err1 := datadogMon.GetUpdatedMetrics(metrics)
	assert.Error(t, err1)

	mockResponse = []datadog.Series{
		{Points: []datadog.DataPoint{}},
	}
	mockDatadogClient.On("QueryMetrics").Return(mockResponse, nil).Once()
	_, err2 := datadogMon.GetUpdatedMetrics(metrics)
	assert.Error(t, err2)
}

func TestDatadogMonitorInvalidApiKey(t *testing.T) {
	setupDatadogMonitorTest()
	metrics := []string{"foo.bar.baz"}
	mockDatadogClient.On("Validate").Return(false, nil).Once()
	_, eA := datadogMon.GetUpdatedMetrics(metrics)
	assert.Error(t, eA)
	mockDatadogClient.On("Validate").Return(true, errors.New("")).Once()
	_, eB := datadogMon.GetUpdatedMetrics(metrics)
	assert.Error(t, eB)
}
