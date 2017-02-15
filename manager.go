package autoscaler

import (
	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/autoscaler/inventory"
	"github.com/notonthehighstreet/autoscaler/monitor"
	"github.com/notonthehighstreet/autoscaler/strategy"
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