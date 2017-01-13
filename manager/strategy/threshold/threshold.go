package threshold

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/autoscaler/manager/inventory"
	"github.com/notonthehighstreet/autoscaler/manager/monitor"
	"github.com/notonthehighstreet/autoscaler/manager/strategy"
	"github.com/spf13/viper"
)

type ThresholdStrategy struct {
	Config *viper.Viper
	// <metric name>: [<lower threshold>, <upper threshold>]
	// Strategy aims to keep the value in the middle but will always recommend scaling up if any metric
	// is above it's upper threshold
	Inventory inventory.Inventory
	Monitor   monitor.Monitor
	log       *logrus.Entry
}

func New(config *viper.Viper, inv inventory.Inventory, mon monitor.Monitor, log *logrus.Entry) (strategy.Strategy, error) {
	return &ThresholdStrategy{Config: config, Inventory: inv, Monitor: mon, log: log}, nil
}

func (p *ThresholdStrategy) Evaluate() (*strategy.Recommendation, error) {
	finalRecommendation := strategy.SCALEDOWN

	var metricNames []string
	for metricName := range p.Config.GetStringMap("thresholds") {
		metricNames = append(metricNames, metricName)
	}
	metricUpdates, err := p.Monitor.GetUpdatedMetrics(metricNames)
	if err != nil {
		return nil, err
	}
	for _, metric := range *metricUpdates {
		var metricRecommendation strategy.Recommendation
		var invert = 1

		metricConfig := p.Config.Sub("thresholds." + metric.Name)
		if metricConfig.GetBool("invert_scaling") {
			invert = -1
		}
		min := metricConfig.GetFloat64("min")
		max := metricConfig.GetFloat64("max")
		switch {
		case metric.CurrentReading < min:
			metricRecommendation = strategy.Recommendation(int(strategy.SCALEDOWN) * invert)
		case metric.CurrentReading >= min && metric.CurrentReading <= max:
			metricRecommendation = strategy.HOLD
		case metric.CurrentReading > max:
			metricRecommendation = strategy.Recommendation(int(strategy.SCALEUP) * invert)
		default:
			return &finalRecommendation, errors.New("Strategy: Something went wrong")
		}
		p.log.Debugf("Metric: %v value: %v min: %v max: %v. Suggests %v.", metric.Name, metric.CurrentReading, min, max, metricRecommendation)
		if finalRecommendation < metricRecommendation { // Worst case scenario wins
			finalRecommendation = metricRecommendation
		}
	}
	p.log.Debugf("Recommending %v as safest option", finalRecommendation)
	return &finalRecommendation, nil

}
