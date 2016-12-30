package inventory

import (
	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
	"strings"
)

type Inventory interface {
	Total() (int, error)
	Increase() error
	Decrease() error
	Status() Status
}

type Status int

const (
	OK Status = iota
	UPDATING
	FAILED
)

// Create a hash for storing the names of registered inventories and their New() methods
// eg {'foo': foo.New(), 'bar': bar.New(), 'baz': baz.New()}
type factoryFunction func(config *viper.Viper, log *logrus.Entry) Inventory

var inventories = make(map[string]factoryFunction)

func Register(name string, factory factoryFunction) {
	if factory == nil {
		logrus.Panicf("New() for %s does not exist.", name)
	}
	_, registered := inventories[name]
	if registered {
		logrus.Errorf("New() for %s already registered. Ignoring.", name)
	}
	inventories[name] = factory
}

func New(config *viper.Viper, log *logrus.Entry) Inventory {
	// Find the correct inventory and return it
	var inv Inventory
	if config.IsSet("name") {
		name := config.GetString("name")
		newFunc, ok := inventories[name]
		if !ok {
			// Inventory has not been registered.
			// Make a list of all available inventories for logging.
			available := make([]string, len(inventories))
			for k := range inventories {
				available = append(available, k)
			}
			log.Fatalf("Invalid inventory name. Must be one of: %s", strings.Join(available, ", "))
		}
		inv = newFunc(config, log.WithField("inventory", name))
	} else {
		// No inventory name provided in config
		log.Fatalf("No inventory name provided")
	}
	return inv
}
