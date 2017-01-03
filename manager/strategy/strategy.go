package strategy

import (
	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/autoscaler/manager/inventory"
	"github.com/notonthehighstreet/autoscaler/manager/monitor"
	"github.com/spf13/viper"
	"strings"
)

type Strategy interface {
	Evaluate() (*Recommendation, error)
}

type Recommendation int

const (
	SCALEDOWN Recommendation = iota - 1
	HOLD
	SCALEUP
)

// Create a hash for storing the names of registered strategies and their New() methods
// eg {'foo': foo.New(), 'bar': bar.New(), 'baz': baz.New()}
type factoryFunction func(config *viper.Viper, inv inventory.Inventory, mon monitor.Monitor, log *logrus.Entry) Strategy

var strategies = make(map[string]factoryFunction)

func Register(name string, factory factoryFunction) {
	if factory == nil {
		logrus.Panicf("New() for %s does not exist.", name)
	}
	_, registered := strategies[name]
	if registered {
		logrus.Errorf("New() for %s already registered. Ignoring.", name)
	}
	strategies[name] = factory
}

func New(config *viper.Viper, inv inventory.Inventory, mon monitor.Monitor, log *logrus.Entry) Strategy {
	// Find the correct monitor and return it
	var str Strategy
	if config.IsSet("name") {
		name := config.GetString("name")
		newFunc, ok := strategies[name]
		if !ok {
			// Strategy has not been registered.
			// Make a list of all available strategies for logging.
			available := make([]string, len(strategies))
			for k := range strategies {
				available = append(available, k)
			}
			log.Fatalf("Invalid strategy name. Must be one of: %s", strings.Join(available, ", "))
		}
		str = newFunc(config, inv, mon, log.WithField("strategy", name))
	} else {
		// No strategy name provided in config
		log.Fatalf("No strategy name provided")
	}
	return str
}
