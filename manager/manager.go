package manager

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/notonthehighstreet/autoscaler/manager/inventory"
	"github.com/notonthehighstreet/autoscaler/manager/inventory/aws"
	"github.com/notonthehighstreet/autoscaler/manager/monitor/mesos"
	"github.com/notonthehighstreet/autoscaler/manager/strategy"
	"github.com/notonthehighstreet/autoscaler/manager/strategy/threshold"
	"github.com/op/go-logging"
)

type Manager struct {
	Inventory inventory.Inventory
	logger    *logging.Logger
	Name      string
	Strategy  strategy.Strategy
}

func New() Manager {
	// TODO: Pull all this stuff from configuration
	name := "EC2InstancesManager"
	mesosUrl := "http://mesos.service.consul:5050/state"
	region := "eu-west-1"
	thresholds := map[string][2]int{
		"mesos.cluster.cpu.percent_used": [2]int{20, 80},
		"mesos.cluster.mem.percent_used": [2]int{20, 80},
	}

	log := logging.MustGetLogger(name)

	log.Info("Initialising inventory")
	s, err := session.NewSession()
	if err != nil {
		log.Errorf("%s", err.Error())
	}
	s.Config.Region = &region
	inv := aws.New(log, autoscaling.New(s), ec2metadata.New(s))

	log.Info("Initialising monitor")
	client, err := mesos.NewMesosClient(mesosUrl)
	if err != nil {
		log.Errorf("%s", err)
	}
	monitor := mesos.New(log, client)

	log.Info("Initialising strategy")
	str := threshold.New(thresholds, inv, monitor)

	return Manager{Name: name, Strategy: str, Inventory: inv, logger: log}
}

func (m *Manager) Run() error {
	m.logger.Info("Executing strategy")
	rec, err := m.Strategy.Evaluate()
	if err == nil {
		switch rec {
		case strategy.SCALEUP:
			m.logger.Info("Scaling up")
			err = m.Inventory.Increase()
		case strategy.HOLD:
			m.logger.Info("Doing nothing")
		case strategy.SCALEDOWN:
			m.logger.Info("Scaling down")
			err = m.Inventory.Decrease()
		default:
			err = errors.New("Unknown recommendation")

		}
	}
	return err

}
