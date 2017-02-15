package strategy

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/autoscaler/manager/inventory"
	"github.com/notonthehighstreet/autoscaler/manager/monitor"
	"github.com/spf13/viper"
)

type Threshold struct {
	Config *viper.Viper
	// <metric name>: [<lower threshold>, <upper threshold>]
	// Strategy aims to keep the value in the middle but will always recommend scaling up if any metric
	// is above it's upper threshold
	Inventory inventory.Inventory
	Monitor   monitor.Monitor
	log       *logrus.Entry
}

func NewThreshold(config *viper.Viper, inv inventory.Inventory, mon monitor.Monitor, log *logrus.Entry) (Strategy, error) {
	return &Threshold{Config: config, Inventory: inv, Monitor: mon, log: log}, nil
}

func (p *Threshold) Evaluate() (*Recommendation, error) {
	finalRecommendation := SCALEDOWN

	var metricNames []string
	for metricName := range p.Config.GetStringMap("thresholds") {
		metricNames = append(metricNames, metricName)
	}
	metricUpdates, err := p.Monitor.GetUpdatedMetrics(metricNames)
	if err != nil {
		return nil, err
	}
	for _, metric := range *metricUpdates {
		var metricRecommendation Recommendation
		var invert = 1

		if !p.Config.IsSet("thresholds." + metric.Name) {
			return nil, fmt.Errorf("No threshold configuration for %s", metric.Name)
		}
		metricConfig := p.Config.Sub("thresholds." + metric.Name)

		if metricConfig.GetBool("invert_scaling") {
			invert = -1
		}
		min := metricConfig.GetFloat64("min")
		max := metricConfig.GetFloat64("max")
		switch {
		case metric.CurrentReading < min && metricConfig.IsSet("min"):
			metricRecommendation = Recommendation(int(SCALEDOWN) * invert)
		case metric.CurrentReading > max && metricConfig.IsSet("max"):
			metricRecommendation = Recommendation(int(SCALEUP) * invert)
		case !metricConfig.IsSet("max") && !metricConfig.IsSet("min"):
			return nil, fmt.Errorf("Threshold strategy needs either 'min' or 'max' for %s", metric.Name)
		default:
			metricRecommendation = HOLD
		}
		p.log.Debugf("Metric: %v value: %v. Suggests %v.", metric.Name, metric.CurrentReading, metricRecommendation)
		if finalRecommendation < metricRecommendation { // Worst case scenario wins
			finalRecommendation = metricRecommendation
		}
	}
	p.log.Debugf("Recommending %v as safest option", finalRecommendation)
	return &finalRecommendation, nil

}
