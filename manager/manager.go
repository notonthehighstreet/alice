package manager

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/autoscaler/manager/inventory"
	"github.com/notonthehighstreet/autoscaler/manager/inventory/aws"
	"github.com/notonthehighstreet/autoscaler/manager/inventory/fake"
	"github.com/notonthehighstreet/autoscaler/manager/monitor"
	"github.com/notonthehighstreet/autoscaler/manager/monitor/datadog"
	"github.com/notonthehighstreet/autoscaler/manager/monitor/fake"
	"github.com/notonthehighstreet/autoscaler/manager/monitor/mesos"
	"github.com/notonthehighstreet/autoscaler/manager/strategy"
	"github.com/notonthehighstreet/autoscaler/manager/strategy/ratio"
	"github.com/notonthehighstreet/autoscaler/manager/strategy/threshold"
	"github.com/spf13/viper"
)

type Manager struct {
	Inventory inventory.Inventory
	Logger    *logrus.Entry
	Strategy  strategy.Strategy
	Config    *viper.Viper
}

func New(config *viper.Viper, log *logrus.Entry) Manager {
	requiredKeys := []string{"inventory", "monitor", "strategy"}

	for _, k := range requiredKeys {
		if !config.IsSet(k) {
			log.Fatalf("Missing %v definition", k)
		}
	}
	log.Info("Initialising inventory")
	inv := inventory.New(config.Sub("inventory"), log)

	log.Info("Initialising monitor")
	monitor := monitor.New(config.Sub("monitor"), log)

	log.Info("Initialising strategy")
	str := strategy.New(config.Sub("strategy"), inv, monitor, log)

	return Manager{Strategy: str, Inventory: inv, Logger: log, Config: config}
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
					m.Logger.Infof("Can't scale up: %v", err.Error())
				} else {
					m.Logger.Warnf("Scaling up our %v inventory based on the %v strategy using information from %v", invName, stratName, monName)
				}
			} else {
				m.Logger.Warn("I would have scaled up")
			}
		case strategy.HOLD:
			m.Logger.Info("Doing nothing")
		case strategy.SCALEDOWN:
			if m.Config.GetBool("scale_down") {
				err = m.Inventory.Decrease()
				if err != nil {
					m.Logger.Infof("Can't scale down: %v", err.Error())
				} else {
					m.Logger.Warnf("Scaling down our %v inventory based on the %v strategy using information from %v", invName, stratName, monName)
				}
			} else {
				m.Logger.Warn("I would have scaled down")
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
	monitor.Register("fake", fake_monitor.New)
	monitor.Register("mesos", mesos.New)
	monitor.Register("datadog", datadog.New)
	strategy.Register("ratio", ratio.New)
	strategy.Register("threshold", threshold.New)
}
