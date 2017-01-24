package newrelic

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/autoscaler/manager/monitor"
	"github.com/spf13/viper"
)

type NewRelicMonitor struct {
	log    *logrus.Entry
	config *viper.Viper
	Client Client
}

type Client interface {
	Get(url string, apiKey string) (NewRelicResponse, error)
}

type NewRelicClient struct{}

type NewRelicResponse struct {
	Application struct {
		ApplicationSummary struct {
			ResponseTime float64 `json:"response_time"`
			Throughput   float64 `json:"throughput"`
			ErrorRate    float64 `json:"error_rate"`
			ApdexScore   float64 `json:"apdex_score"`
		} `json:"application_summary"`
	} `json:"application"`
}

func (r *NewRelicResponse) GetMetric(name string) float64 {
	app := r.Application.ApplicationSummary
	mapping := map[string]float64{
		"response_time": app.ResponseTime,
		"throughput":    app.Throughput,
		"error_rate":    app.ErrorRate,
		"apdex_score":   app.ApdexScore,
	}
	return mapping[name]
}

func New(config *viper.Viper, log *logrus.Entry) (monitor.Monitor, error) {
	requiredConfig := []string{"api_key", "app_id"}
	for _, item := range requiredConfig {
		if !config.IsSet(item) {
			log.Fatalf("NewRelic config missing: %v", item)
		}
	}
	return NewRelicMonitor{config: config, log: log, Client: &NewRelicClient{}}, nil
}

func (nr NewRelicMonitor) GetUpdatedMetrics(names []string) (*[]monitor.MetricUpdate, error) {
	endpoint := fmt.Sprintf("https://api.newrelic.com/v2/applications/%s.json", nr.config.GetString("app_id"))
	newrelicResponse, err := nr.Client.Get(endpoint, nr.config.GetString("api_key"))
	if err != nil {
		return nil, err
	}

	metrics := make([]monitor.MetricUpdate, len(names))
	for i, name := range names {
		metrics[i] = monitor.MetricUpdate{Name: name, CurrentReading: newrelicResponse.GetMetric(name)}
	}
	return &metrics, nil
}

func (c *NewRelicClient) Get(URL, apiKey string) (NewRelicResponse, error) {
	var nrResponse NewRelicResponse

	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return nrResponse, err
	}
	req.Header.Add("X-Api-Key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nrResponse, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nrResponse, err
	}
	if err = json.Unmarshal(body, &nrResponse); err != nil {
		return nrResponse, err
	}
	return nrResponse, nil
}
