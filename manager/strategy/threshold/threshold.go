package threshold

import (
	"errors"
	"github.com/notonthehighstreet/autoscaler/manager/inventory"
	"github.com/notonthehighstreet/autoscaler/manager/monitor"
	"github.com/notonthehighstreet/autoscaler/manager/strategy"
	"github.com/sirupsen/logrus"
)

type ThresholdStrategy struct {
	Thresholds map[string][2]int
	// <metric name>: [<lower threshold>, <upper threshold>]
	// Strategy aims to keep the value in the middle but will always recommend scaling up if any metric
	// is above it's upper threshold
	Inventory inventory.Inventory
	Monitor   monitor.Monitor
	logger    *logrus.Entry
}

func New(thr map[string][2]int, inv inventory.Inventory, mon monitor.Monitor, logger *logrus.Entry) *ThresholdStrategy {
	return &ThresholdStrategy{Thresholds: thr, Inventory: inv, Monitor: mon, logger: logger}
}

func (p *ThresholdStrategy) Evaluate() (strategy.Recommendation, error) {
	finalRecommendation := strategy.SCALEDOWN

	var metricNames []string
	for metricName, _ := range p.Thresholds {
		metricNames = append(metricNames, metricName)
	}
	metricUpdates, err := p.Monitor.GetUpdatedMetrics(metricNames)
	if err != nil {
		return finalRecommendation, err
	}
	for _, metric := range *metricUpdates {
		var metricRecommendation strategy.Recommendation
		switch {
		case metric.CurrentReading < p.Thresholds[metric.Name][0]:
			metricRecommendation = strategy.SCALEDOWN
		case metric.CurrentReading >= p.Thresholds[metric.Name][0] && metric.CurrentReading <= p.Thresholds[metric.Name][1]:
			metricRecommendation = strategy.HOLD
		case metric.CurrentReading > p.Thresholds[metric.Name][1]:
			metricRecommendation = strategy.SCALEUP
		default:
			return finalRecommendation, errors.New("Strategy: Something went wrong")
		}
		if finalRecommendation < metricRecommendation { // Worst case scenario wins
			finalRecommendation = metricRecommendation
		}
	}
	return finalRecommendation, nil

}
