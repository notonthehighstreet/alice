package alice

import (
	"errors"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
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
type strategyFactoryFunc func(config *viper.Viper, inv Inventory, mon Monitor, log *logrus.Entry) (Strategy, error)

var strategies = make(map[string]strategyFactoryFunc)

func RegisterStrategy(name string, factory strategyFactoryFunc) {
	if factory == nil {
		logrus.Panicf("New() for %s does not exist.", name)
	}
	_, registered := strategies[name]
	if registered {
		logrus.Errorf("New() for %s already registered. Ignoring.", name)
	}
	strategies[name] = factory
}

func NewStrategy(config *viper.Viper, inv Inventory, mon Monitor, log *logrus.Entry) (Strategy, error) {
	// Find the correct monitor and return it
	if !config.IsSet("name") {
		return nil, errors.New("No strategy name provided")
	}
	name := config.GetString("name")
	newFunc, ok := strategies[name]
	if !ok {
		available := make([]string, len(strategies))
		for s := range strategies {
			available = append(available, s)
		}
		log.Fatalf("Invalid strategy name. Must be one of: %s", strings.Join(available, ", "))
	}
	return newFunc(config, inv, mon, log.WithField("strategy", name))
}
