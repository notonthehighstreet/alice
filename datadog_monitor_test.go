package autoscaler_test

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/autoscaler"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	datadogclient "github.com/zorkian/go-datadog-api"
	"testing"
	"time"
)

var datadogMon *autoscaler.DatadogMonitor
var mockDatadogClient autoscaler.MockDatadogClient

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
	mockDatadogClient = autoscaler.MockDatadogClient{}
	m, _ := autoscaler.NewDatadogMonitor(config, log)
	datadogMon = m.(*autoscaler.DatadogMonitor)
	datadogMon.Client = &mockDatadogClient
}

func TestDatadogMonitor_GetUpdatedMetrics(t *testing.T) {
	setupDatadogMonitorTest()
	config.Set("metrics.foo.bar.baz.query", "avg:foo.bar.baz{*}")
	metrics := []string{"foo.bar.baz"}
	mockResponse := []datadogclient.Series{
		{Points: []datadogclient.DataPoint{
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

	mockResponse := []datadogclient.Series{}
	mockDatadogClient.On("QueryMetrics").Return(mockResponse, nil).Once()
	_, err1 := datadogMon.GetUpdatedMetrics(metrics)
	assert.Error(t, err1)

	mockResponse = []datadogclient.Series{
		{Points: []datadogclient.DataPoint{}},
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