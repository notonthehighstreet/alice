package alice

import (
	"errors"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
)

// Inventory represents the generic inventory interface. An inventory can manage any type of resource (server instances,
// application instances, sheep etc). As long as it can return a total, be scaled up and down, and let us know if the
// inventory is healthy, then it will work..
type Inventory interface {
	Total() (int, error)
	Increase() error
	Decrease() error
	Status() (Status, error)
}

// Status represents the various statuses that can be returned by an inventory's Status() function.
type Status int

const (
	// OK - means the inventory is ready to be scaled and we can call Increase/Decrease.
	OK Status = iota
	// UPDATING - means the inventory is still making changes after the last scaling action and we should wait.
	UPDATING
	// FAILED - means something is wrong, and we should definitely not trust the state of the inventory or change anything
	FAILED
)

// Create a hash for storing the names of registered inventories and their New() methods
// eg {'foo': foo.New(), 'bar': bar.New(), 'baz': baz.New()}
type inventoryFactoryFunc func(config *viper.Viper, log *logrus.Entry) (Inventory, error)

var inventories = make(map[string]inventoryFactoryFunc)

// RegisterInventory allows a new inventory type to be registered with a string name. This name is used to match
// configuration to the correct NewFooInventory function that can read it.
func RegisterInventory(name string, factory inventoryFactoryFunc) {
	if factory == nil {
		logrus.Panicf("New() for %s does not exist.", name)
	}
	_, registered := inventories[name]
	if registered {
		logrus.Errorf("New() for %s already registered. Ignoring.", name)
	}
	inventories[name] = factory
}

// NewInventory will take a generic block of configuration and read look for a 'name' key, and immediately pass
// the block of config to the factory function that has been registered with that name.
func NewInventory(config *viper.Viper, log *logrus.Entry) (Inventory, error) {
	// Find the correct inventory and return it
	if !config.IsSet("name") {
		return nil, errors.New("No inventory name provided")
	}
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
	return newFunc(config, log.WithField("inventory", name))

}
