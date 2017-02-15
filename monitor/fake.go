package monitor

import (
	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
	"math"
)

type Fake struct {
	log       *logrus.Entry
	config    *viper.Viper
	iteration int
}

func (f *Fake) GetUpdatedMetrics(names []string) (*[]MetricUpdate, error) {
	response := make([]MetricUpdate, len(names))
	fakeReading := f.generateFakeReading()
	f.log.Infof("Setting all metrics to the fake reading %v", fakeReading)
	for i, name := range names {
		response[i].Name = name
		response[i].CurrentReading = float64(fakeReading)
	}
	return &response, nil
}

func (f *Fake) generateFakeReading() int {
	// Fake a reading. At the moment just generating a sine wave to simulate a metric that rises and falls.
	input := float64(f.iteration*f.config.GetInt("increments")) * math.Pi / 180
	output := (math.Sin(input) + 1) * 50
	f.iteration += 1
	return int(output)
}

func NewFake(config *viper.Viper, log *logrus.Entry) (Monitor, error) {
	config.SetDefault("increments", 10)
	return &Fake{config: config, log: log, iteration: 0}, nil
}
