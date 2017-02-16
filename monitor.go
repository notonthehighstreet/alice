package alice

import (
	"errors"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
)

// Monitor represents the generic Monitor interface. Monitors are a source of information that feed in to strategies.
// Given a list of metrics (often in dot-notation for a lot of providers, eg sys.mem.free), a monitor must provide
// a current reading for each one.
type Monitor interface {
	GetUpdatedMetrics([]string) (*[]MetricUpdate, error)
}

// MetricUpdate stores the name of the metric requested and its current reading.
type MetricUpdate struct {
	Name           string
	CurrentReading float64
}

// Create a hash for storing the names of registered monitors and their New() methods
// eg {'foo': foo.New(), 'bar': bar.New(), 'baz': baz.New()}
type monitorFactoryFunc func(config *viper.Viper, log *logrus.Entry) (Monitor, error)

var monitors = make(map[string]monitorFactoryFunc)

// RegisterMonitor allows a new monitor type to be registered with a string name. This name is used to match
// configuration to the correct NewFooMonitor function that can read it.
func RegisterMonitor(name string, factory monitorFactoryFunc) {
	if factory == nil {
		logrus.Panicf("New() for %s does not exist.", name)
	}
	_, registered := monitors[name]
	if registered {
		logrus.Errorf("New() for %s already registered. Ignoring.", name)
	}
	monitors[name] = factory
}

// NewMonitor will take a generic block of configuration and read look for a 'name' key, and immediately pass
// the block of config to the factory function that has been registered with that name.
func NewMonitor(config *viper.Viper, log *logrus.Entry) (Monitor, error) {
	// Find the correct monitor and return it
	if !config.IsSet("name") {
		return nil, errors.New("No monitor name provided")
	}

	name := config.GetString("name")
	newFunc, ok := monitors[name]
	if !ok {
		available := make([]string, len(monitors))
		for m := range monitors {
			available = append(available, m)
		}
		log.Fatalf("Invalid monitor name. Must be one of: %s", strings.Join(available, ", "))
	}
	return newFunc(config, log.WithField("monitor", name))
}
