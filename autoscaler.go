package main

import (
	"time"

	"github.com/notonthehighstreet/autoscaler/mesos"
	"github.com/notonthehighstreet/autoscaler/scaler"

	"github.com/op/go-logging"
)

func main() {
	var log = logging.MustGetLogger("autoscaler")

	// Get the Mesos URL from Consul
	mesosMaster := mesos.NewMesosMaster(log)

	log.Info("Running autoscaler")
	scaler, err := scaler.NewScaler(log)
	if err != nil {
		log.Errorf("scaler: %s", err.Error())
	}

	for _ = range time.NewTicker(2 * time.Second).C {
		stats, err := mesosMaster.Stats()
		if err != nil {
			log.Errorf("mesos: %s", err.Error())
		}

		if stats.CPUPercent >= 0.8 {
			err := scaler.ScaleUp()
			if err != nil {
				log.Errorf("Error scaling up - %v", err)
			}
			log.Info("Scaled up by one due to CPU pressure!")

		} else if stats.MemPercent >= 0.8 {
			err := scaler.ScaleUp()
			if err != nil {
				log.Errorf("Error scaling up - %v", err)
			}
			log.Info("Scaled up by one due to memory pressure!")
		}

	}
}
