package newrelic_test

import (
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/autoscaler/manager/monitor/newrelic"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var log = logrus.WithFields(logrus.Fields{
	"manager": "Mock",
	"monitor": "NewRelicMonitor",
})
var monitor newrelic.NewRelicMonitor
var config *viper.Viper
var mockNewRelicClient newrelic.MockNewRelicClient

func setupTest() {
	config = viper.New()
	config.Set("api_key", "foo")
	config.Set("app_id", "bar")
	config.Set("time_period", "5m")

	log.Logger.Level = logrus.DebugLevel
	mockNewRelicClient = newrelic.MockNewRelicClient{}
	m, _ := newrelic.New(config, log)
	monitor = m.(newrelic.NewRelicMonitor)
	monitor.Client = &mockNewRelicClient
}

func TestNewRelic_GetUpdatedMetrics(t *testing.T) {
	setupTest()
	config.Set("metric", "something")
	response := newrelic.NewRelicResponse{}
	response.Application.ApplicationSummary.ApdexScore = 20.0
	response.Application.ApplicationSummary.Throughput = 20.0
	response.Application.ApplicationSummary.ErrorRate = 20.0
	response.Application.ApplicationSummary.ApdexScore = 20.0

	mockNewRelicClient.
		On("Get", "https://api.newrelic.com/v2/applications/bar.json", "foo").
		Return(response, nil)

	mts, err := monitor.GetUpdatedMetrics([]string{"throughput"})
	assert.NoError(t, err)
	val := *mts
	assert.Equal(t, 1, len(val))
	assert.Equal(t, 20.0, val[0].CurrentReading)
}

func TestNewRelic_GetUpdatedMetricsNoData(t *testing.T) {
	setupTest()
	config.Set("metric", "something")
	response := newrelic.NewRelicResponse{}

	mockNewRelicClient.
		On("Get", "https://api.newrelic.com/v2/applications/bar.json", "foo").
		Return(response, nil)

	mts, err := monitor.GetUpdatedMetrics([]string{"throughput"})
	assert.NoError(t, err)
	val := *mts
	assert.Equal(t, 1, len(val))
	assert.Equal(t, 0.0, val[0].CurrentReading)

}
