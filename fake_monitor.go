package alice

import (
	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
	"math"
)

// FakeMonitor is a dummy monitor that will generate numbers for any metric requested. Values returned for metrics
// are between 1 and 100 and follow a sine-wave pattern which is useful for testing demand that gradually changes with
// time.
type FakeMonitor struct {
	log       *logrus.Entry
	config    *viper.Viper
	iteration int
}

// GetUpdatedMetrics returns MetricUpdates for each of the metrics requested
func (f *FakeMonitor) GetUpdatedMetrics(names []string) (*[]MetricUpdate, error) {
	response := make([]MetricUpdate, len(names))
	fakeReading := f.generateFakeReading()
	f.log.Infof("Setting all metrics to the fake reading %v", fakeReading)
	for i, name := range names {
		response[i].Name = name
		response[i].CurrentReading = float64(fakeReading)
	}
	return &response, nil
}

func (f *FakeMonitor) generateFakeReading() int {
	// Fake a reading. At the moment just generating a sine wave to simulate a metric that rises and falls.
	input := float64(f.iteration*f.config.GetInt("increments")) * math.Pi / 180
	output := (math.Sin(input) + 1) * 50
	f.iteration++
	return int(output)
}

// NewFakeMonitor returns a new Monitor
func NewFakeMonitor(config *viper.Viper, log *logrus.Entry) (Monitor, error) {
	config.SetDefault("increments", 10)
	return &FakeMonitor{config: config, log: log, iteration: 0}, nil
}
