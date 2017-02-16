package autoscaler

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
	"strings"
)

type Monitor interface {
	GetUpdatedMetrics([]string) (*[]MetricUpdate, error)
}

type MetricUpdate struct {
	Name           string
	CurrentReading float64
}

// Create a hash for storing the names of registered monitors and their New() methods
// eg {'foo': foo.New(), 'bar': bar.New(), 'baz': baz.New()}
type monitorFactoryFunc func(config *viper.Viper, log *logrus.Entry) (Monitor, error)

var monitors = make(map[string]monitorFactoryFunc)

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
