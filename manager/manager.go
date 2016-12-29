package manager

import (
	"errors"
	"github.com/notonthehighstreet/autoscaler/manager/inventory"
	"github.com/notonthehighstreet/autoscaler/manager/inventory/aws"
	"github.com/notonthehighstreet/autoscaler/manager/inventory/fake"
	"github.com/notonthehighstreet/autoscaler/manager/monitor"
	"github.com/notonthehighstreet/autoscaler/manager/monitor/fake"
	"github.com/notonthehighstreet/autoscaler/manager/monitor/mesos"
	"github.com/notonthehighstreet/autoscaler/manager/strategy"
	"github.com/notonthehighstreet/autoscaler/manager/strategy/threshold"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Manager struct {
	Inventory inventory.Inventory
	logger    *logrus.Entry
	Strategy  strategy.Strategy
}

func New(config *viper.Viper, log *logrus.Entry) Manager {
	log.Info("Initialising inventory")
	inv := inventory.New(config.Sub("inventory"), log)

	log.Info("Initialising monitor")
	monitor := monitor.New(config.Sub("monitor"), log)

	log.Info("Initialising strategy")
	str := threshold.New(config.Sub("strategy"), inv, monitor, log.WithField("strategy", "ThresholdStrategy"))

	return Manager{Strategy: str, Inventory: inv, logger: log}
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

func init() {
	// Register plugins at load time
	inventory.Register("aws", aws.New)
	inventory.Register("fake", fake_inventory.New)
	monitor.Register("fake", fake_monitor.New)
	monitor.Register("mesos", mesos.New)

}
