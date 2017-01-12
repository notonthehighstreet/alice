package manager

import (
	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/autoscaler/manager/inventory"
	"github.com/notonthehighstreet/autoscaler/manager/inventory/aws"
	"github.com/notonthehighstreet/autoscaler/manager/inventory/fake"
	"github.com/notonthehighstreet/autoscaler/manager/inventory/marathon"
	"github.com/notonthehighstreet/autoscaler/manager/monitor"
	"github.com/notonthehighstreet/autoscaler/manager/monitor/datadog"
	"github.com/notonthehighstreet/autoscaler/manager/monitor/fake"
	"github.com/notonthehighstreet/autoscaler/manager/monitor/mesos"
	"github.com/notonthehighstreet/autoscaler/manager/strategy"
	"github.com/notonthehighstreet/autoscaler/manager/strategy/ratio"
	"github.com/notonthehighstreet/autoscaler/manager/strategy/threshold"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Manager struct {
	Inventory inventory.Inventory
	Logger    *logrus.Entry
	Strategy  strategy.Strategy
	Config    *viper.Viper
}

func New(config *viper.Viper, log *logrus.Entry) (Manager, error) {
	requiredKeys := []string{"inventory", "monitor", "strategy"}
	for _, k := range requiredKeys {
		if !config.IsSet(k) {
			log.Fatalf("Missing %v definition", k)
		}
	}

	log.Info("Initialising inventory")
	inv, err := inventory.New(config.Sub("inventory"), log)
	if err != nil {
		return Manager{}, errors.Wrap(err, "Error initialization inventory")
	}

	log.Info("Initialising monitor")
	monitor, err := monitor.New(config.Sub("monitor"), log)
	if err != nil {
		return Manager{}, errors.Wrap(err, "Error initialization monitor")
	}

	log.Info("Initialising strategy")
	str, err := strategy.New(config.Sub("strategy"), inv, monitor, log)
	if err != nil {
		return Manager{}, errors.Wrap(err, "Error initializing strategy")
	}

	return Manager{Strategy: str, Inventory: inv, Logger: log, Config: config}, nil
}

func (m *Manager) Run() error {
	m.Logger.Info("Executing strategy")
	rec, err := m.Strategy.Evaluate()
	m.Config.SetDefault("scale_up", true)
	m.Config.SetDefault("scale_down", true)
	invName, stratName, monName := m.Config.GetString("inventory.name"), m.Config.GetString("strategy.name"), m.Config.GetString("monitor.name")
	if err == nil {
		switch *rec {
		case strategy.SCALEUP:
			if m.Config.GetBool("scale_up") {
				err = m.Inventory.Increase()
				if err != nil {
					m.Logger.Infof("Can't scale up: %s", err.Error())
				} else {
					m.Logger.Warnf("Scaling up our %s inventory based on the %s strategy using information from %s", invName, stratName, monName)
				}
			} else {
				m.Logger.Warnf("I would have scaled up our %s inventory based on the %s strategy using information from %s but am running in advisory mode", invName, stratName, monName)
			}
		case strategy.HOLD:
			m.Logger.Info("Doing nothing")
		case strategy.SCALEDOWN:
			if m.Config.GetBool("scale_down") {
				err = m.Inventory.Decrease()
				if err != nil {
					m.Logger.Infof("Can't scale down: %s", err.Error())
				} else {
					m.Logger.Warnf("Scaling down our %s inventory based on the %s strategy using information from %s", invName, stratName, monName)
				}
			} else {
				m.Logger.Warnf("I would have scaled down our %s inventory based on the %s strategy using information from %s but am running in advisory mode", invName, stratName, monName)
			}
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
	inventory.Register("marathon", marathon.New)
	monitor.Register("fake", fake_monitor.New)
	monitor.Register("mesos", mesos.New)
	monitor.Register("datadog", datadog.New)
	strategy.Register("ratio", ratio.New)
	strategy.Register("threshold", threshold.New)
}
