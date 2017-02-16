package alice

import (
	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// Manager ties together the inventory and the strategy. It will evaluate the strategy and will execute scaling actions
// on the inventory based on the recommendation it received.
type Manager struct {
	Inventory Inventory
	Logger    *logrus.Entry
	Strategy  Strategy
	Config    *viper.Viper
}

// New creates a new Manager
func New(config *viper.Viper, log *logrus.Entry) (Manager, error) {
	requiredKeys := []string{"inventory", "monitor", "strategy"}
	for _, k := range requiredKeys {
		if !config.IsSet(k) {
			log.Fatalf("Missing %v definition", k)
		}
	}

	log.Info("Initialising inventory")
	inv, err := NewInventory(config.Sub("inventory"), log)
	if err != nil {
		return Manager{}, errors.Wrap(err, "Error initialization inventory")
	}

	log.Info("Initialising monitor")
	monitor, err := NewMonitor(config.Sub("monitor"), log)
	if err != nil {
		return Manager{}, errors.Wrap(err, "Error initialization monitor")
	}

	log.Info("Initialising strategy")
	str, err := NewStrategy(config.Sub("strategy"), inv, monitor, log)
	if err != nil {
		return Manager{}, errors.Wrap(err, "Error initializing strategy")
	}

	return Manager{Strategy: str, Inventory: inv, Logger: log, Config: config}, nil
}

// Run requests a recommendation from the strategy, and if not running in dry-run mode will attempt to scale up the
// inventory.
func (m *Manager) Run() error {
	m.Logger.Info("Executing strategy")
	rec, err := m.Strategy.Evaluate()
	m.Config.SetDefault("scale_up", true)
	m.Config.SetDefault("scale_down", true)
	invName, stratName, monName := m.Config.GetString("inventory.name"), m.Config.GetString("strategy.name"), m.Config.GetString("monitor.name")
	if err == nil {
		switch *rec {
		case SCALEUP:
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
		case HOLD:
			m.Logger.Info("Doing nothing")
		case SCALEDOWN:
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
