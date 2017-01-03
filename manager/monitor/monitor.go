package monitor

import (
	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
	"strings"
)

type Monitor interface {
	GetUpdatedMetrics([]string) (*[]MetricUpdate, error)
}

type MetricUpdate struct {
	Name           string
	CurrentReading int
}

// Create a hash for storing the names of registered monitors and their New() methods
// eg {'foo': foo.New(), 'bar': bar.New(), 'baz': baz.New()}
type factoryFunction func(config *viper.Viper, log *logrus.Entry) Monitor

var monitors = make(map[string]factoryFunction)

func Register(name string, factory factoryFunction) {
	if factory == nil {
		logrus.Panicf("New() for %s does not exist.", name)
	}
	_, registered := monitors[name]
	if registered {
		logrus.Errorf("New() for %s already registered. Ignoring.", name)
	}
	monitors[name] = factory
}

func New(config *viper.Viper, log *logrus.Entry) Monitor {
	// Find the correct monitor and return it
	var mon Monitor
	if config.IsSet("name") {
		name := config.GetString("name")
		newFunc, ok := monitors[name]
		if !ok {
			// Monitor has not been registered.
			// Make a list of all available monitors for logging.
			available := make([]string, len(monitors))
			for k := range monitors {
				available = append(available, k)
			}
			log.Fatalf("Invalid monitor name. Must be one of: %s", strings.Join(available, ", "))
		}
		mon = newFunc(config, log.WithField("monitor", name))
	} else {
		// No monitor name provided in config
		log.Fatalf("No monitor name provided")
	}
	return mon
}
