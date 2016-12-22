package main

import (
	"time"

	"github.com/notonthehighstreet/autoscaler/mesos"
	"github.com/notonthehighstreet/autoscaler/scaler"

	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/op/go-logging"
)

var MesosURL = "http://mesos.service.consul:5050/state"

func main() {
	var log = logging.MustGetLogger("autoscaler")

	// Get the Mesos URL from Consul
	client, err := mesos.NewMesosClient(MesosURL)
	if err != nil {
		log.Errorf("mesos: %s", err)
	}
	mesosMaster := mesos.NewMesosMaster(log, client)

	log.Info("Running autoscaler")

	// create our session to AWS
	s, err := session.NewSession()
	if err != nil {
		log.Errorf("%s", err.Error())
	}

	scaler, err := scaler.NewScaler(log, autoscaling.New(s), ec2metadata.New(s))
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
