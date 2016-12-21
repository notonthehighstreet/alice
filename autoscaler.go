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
	mesosMaster := mesos.NewMesosMaster(log, mesos.NewMesosClient("http://mesos.service.consul:5050/state"))

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
		stats.LogUsage(log)

		if stats.CPUPercent >= 0.8 || stats.MemPercent >= 0.8 {
			err := scaler.Scale(1)
			if err != nil {
				log.Errorf("Error scaling up - %v", err)
			}
			log.Info("Scaled up by one due to CPU/Mem pressure!")
		} else if stats.CPUPercent <= 0.2 || stats.MemPercent <= 0.2 {
			err := scaler.Scale(-1)
			if err != nil {
				log.Errorf("Error scaling down - %v", err)
			}
			log.Info("Scaled down by one due to CPU/Mem pressure!")
		}
	}
}
