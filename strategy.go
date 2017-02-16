package alice

import (
	"errors"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
)

// Strategy represents the generic Strategy interface. A strategy contains the logic required to take some information
// and come to a decision about whether an inventory of resources needs to be increased or decreased.
// It doesn't make any changes, and only provides a recommendation to a Manager.
type Strategy interface {
	Evaluate() (*Recommendation, error)
}

// Recommendation is the return type representing the action the strategy recommends the Manager take
type Recommendation int

const (
	// SCALEDOWN - we have too much inventory, decrease it
	SCALEDOWN Recommendation = iota - 1
	// HOLD - the inventory is just right
	HOLD
	// SCALEUP - we have too little inventory, increase it
	SCALEUP
)

// Create a hash for storing the names of registered strategies and their New() methods
// eg {'foo': foo.New(), 'bar': bar.New(), 'baz': baz.New()}
type strategyFactoryFunc func(config *viper.Viper, inv Inventory, mon Monitor, log *logrus.Entry) (Strategy, error)

var strategies = make(map[string]strategyFactoryFunc)

// RegisterStrategy allows a new strategy type to be registered with a string name. This name is used to match
// configuration to the correct NewFooStrategy function that can read it.
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

// NewStrategy will take a generic block of configuration and read look for a 'name' key, and immediately pass
// the block of config to the factory function that has been registered with that name.
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
