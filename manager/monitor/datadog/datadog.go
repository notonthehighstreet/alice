package datadog

import (
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/autoscaler/manager/monitor"
	"github.com/spf13/viper"
	"github.com/zorkian/go-datadog-api"
	"sort"
	"strconv"
	"strings"
	"time"
)

type DatadogMonitorClient interface {
	Validate() (bool, error)
	QueryMetrics(int64, int64, string) ([]datadog.Series, error)
}

type DatadogMonitor struct {
	log    *logrus.Entry
	Client DatadogMonitorClient
	config *viper.Viper
}

func (d *DatadogMonitor) GetUpdatedMetrics(names []string) (*[]monitor.MetricUpdate, error) {
	response := make([]monitor.MetricUpdate, len(names))
	to := time.Now()
	from := to.Add(-d.config.GetDuration("time_period"))
	for index, metric := range names {
		query, err := d.BuildQuery(metric)
		if err != nil {
			return nil, err
		}
		valid, err := d.Client.Validate() // validate API key
		if err != nil {
			return nil, err
		}
		if !valid {
			return nil, errors.New("Datadog API key invalid")
		}
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

func (d *DatadogMonitor) BuildQuery(metric string) (string, error) {
	var query string
	var tags = "{*}"
	d.config.SetDefault("aggregation_method", "avg")
	validMethods := []string{"avg", "max", "min", "sum"}
	agg := d.config.GetString("aggregation_method")
	if !stringInSlice(agg, validMethods) {
		return query, errors.New("Invalid aggregation method")
	}
	if d.config.IsSet("tags") {
		tags = d.BuildTagsString(d.config.GetStringMapString("tags"))
	}
	return agg + ":" + metric + tags, nil
}

func (d *DatadogMonitor) BuildTagsString(tags map[string]string) string {
	var tagsArray []string
	for key, value := range tags {
		tagsArray = append(tagsArray, key+":"+value)
	}
	sort.Strings(tagsArray)
	return "{" + strings.Join(tagsArray, ",") + "}"
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func New(config *viper.Viper, log *logrus.Entry) (monitor.Monitor, error) {
	requiredConfig := []string{"api_key", "app_key", "time_period"}
	for _, item := range requiredConfig {
		if !config.IsSet(item) {
			log.Fatalf("Missing config: %v", item)
		}
	}
	client := datadog.NewClient(config.GetString("api_key"), config.GetString("app_key"))
	return &DatadogMonitor{log: log, config: config, Client: client}, nil
}
