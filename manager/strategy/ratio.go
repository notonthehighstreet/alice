package strategy

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/autoscaler/manager/inventory"
	"github.com/notonthehighstreet/autoscaler/manager/monitor"
	"github.com/spf13/viper"
	"math"
)

// RatioStrategy tries to keep the resources in an inventory at a set ratio to a current metric reading
type Ratio struct {
	Config    *viper.Viper
	Inventory inventory.Inventory
	Monitor   monitor.Monitor
	log       *logrus.Entry
}

func NewRatio(config *viper.Viper, inv inventory.Inventory, mon monitor.Monitor, log *logrus.Entry) (Strategy, error) {
	return &Ratio{Config: config, Inventory: inv, Monitor: mon, log: log}, nil
}

func (r *Ratio) Evaluate() (*Recommendation, error) {
	finalRecommendation := SCALEDOWN

	var metricNames []string
	for metricName := range r.Config.GetStringMap("ratios") {
		metricNames = append(metricNames, metricName)
	}
	metricUpdates, err := r.Monitor.GetUpdatedMetrics(metricNames)
	if err != nil {
		return nil, err
	}
	currentTotal, err := r.Inventory.Total()
	if err != nil {
		return nil, err
	}
	for _, metric := range *metricUpdates {
		metricConfig := r.Config.Sub("ratios." + metric.Name)
		var metricRecommendation Recommendation
		if !metricConfig.IsSet("metric") || !metricConfig.IsSet("inventory") {
			return nil, errors.New("Strategy requires 'metric' and 'inventory' numbers for each ratio")
		}

		m := float64(metricConfig.GetInt("metric"))
		i := float64(metricConfig.GetInt("inventory"))
		c := float64(metric.CurrentReading)
		// Desired state is m/i = c/t. Therefore we should scale t until t = ci/m.
		// Eg if config says metric to inventory should be 3/2, and the current reading is 9, then total
		// inventory should be 9*2/3 = 6. Always round UP to nearest integer.
		t := int(math.Ceil(c * i / m))

		switch {
		case currentTotal < t:
			metricRecommendation = SCALEUP
		case currentTotal == t:
			metricRecommendation = HOLD
		case currentTotal > t:
			metricRecommendation = SCALEDOWN
		default:
			return nil, errors.New("Strategy: Something went wrong")
		}
		r.log.Debugf("Metric: %v value: %v desired metric/inventory ratio: %v/%v. Suggests %v.", metric.Name, metric.CurrentReading, m, i, metricRecommendation)
		if finalRecommendation < metricRecommendation { // Worst case scenario wins
			finalRecommendation = metricRecommendation
		}
	}
	r.log.Debugf("Recommending %v as safest option", finalRecommendation)
	return &finalRecommendation, nil
}