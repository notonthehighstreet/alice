package alice

import (
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/zorkian/go-datadog-api"
	"strconv"
	"time"
)

// DatadogMonitorClient is an intenface allowing mocks of go-datadog-api
type DatadogMonitorClient interface {
	Validate() (bool, error)
	QueryMetrics(int64, int64, string) ([]datadog.Series, error)
}

// DatadogMonitor is a monitor that can pull metrics from Datadog
type DatadogMonitor struct {
	log               *logrus.Entry
	Client            DatadogMonitorClient
	config            *viper.Viper
	isAPIKeyValid     bool
	haveCheckedAPIKey bool
}

// GetUpdatedMetrics returns MetricUpdates for each of the metrics requested
func (d *DatadogMonitor) GetUpdatedMetrics(names []string) (*[]MetricUpdate, error) {
	response := make([]MetricUpdate, len(names))
	if !d.haveCheckedAPIKey {
		v, err := d.Client.Validate() // validate API key
		if err != nil {
			return nil, err
		}
		d.isAPIKeyValid = v
	}
	if !d.isAPIKeyValid {
		return nil, errors.New("Datadog API key invalid")
	}
	to := time.Now()
	from := to.Add(-d.config.GetDuration("time_period"))
	if !d.config.IsSet("metrics") {
		return nil, errors.New("Must provide 'metrics' config")
	}
	metricConfig := d.config.Sub("metrics")
	for index, metric := range names {
		if !metricConfig.IsSet(metric + ".query") {
			return nil, fmt.Errorf("Must provide datadog query for metric %s", metric)
		}
		query := metricConfig.GetString(metric + ".query")
		d.log.Debugln("/v1/query?from=" + strconv.FormatInt(from.Unix(), 10) + "&to=" + strconv.FormatInt(to.Unix(), 10) + "&query=" + query)
		result, err := d.Client.QueryMetrics(from.Unix(), to.Unix(), query)
		d.log.Debugf("Number of items returned: %v", len(result))
		if err != nil {
			d.log.Debug(err.Error())
			return nil, err
		}
		if len(result) != 1 || len(result[0].Points) < 1 {
			return nil, fmt.Errorf("No data for %v between %v and %v", metric, from, to)
		}
		response[index].Name = metric
		response[index].CurrentReading = result[0].Points[len(result[0].Points)-1][1]
	}
	return &response, nil
}

// NewDatadogMonitor returns a new DatadogMonitor
func NewDatadogMonitor(config *viper.Viper, log *logrus.Entry) (Monitor, error) {
	requiredConfig := []string{"api_key", "app_key", "time_period"}
	for _, item := range requiredConfig {
		if !config.IsSet(item) {
			log.Fatalf("Missing config: %v", item)
		}
	}
	client := datadog.NewClient(config.GetString("api_key"), config.GetString("app_key"))
	return &DatadogMonitor{log: log, config: config, Client: client}, nil
}
