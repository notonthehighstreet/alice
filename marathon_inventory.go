package autoscaler

import (
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/gambol99/go-marathon"
	"github.com/spf13/viper"
	"time"
)

type MarathonInventoryClient interface {
	ApplicationBy(name string, opts *marathon.GetAppOpts) (*marathon.Application, error)
	ScaleApplicationInstances(name string, instances int, force bool) (*marathon.DeploymentID, error)
}

type MarathonInventory struct {
	log          *logrus.Entry
	Client       MarathonInventoryClient
	Config       *viper.Viper
	lastModified time.Time
}

func NewMarathonInventory(config *viper.Viper, log *logrus.Entry) (Inventory, error) {
	requiredConfig := []string{"app", "url"}
	for _, item := range requiredConfig {
		if !config.IsSet(item) {
			return nil, errors.New(fmt.Sprintf("Missing config: %v", item))
		}
	}
	config.SetDefault("settle_down_period", "0s")
	marathonConfig := marathon.NewDefaultConfig()
	marathonConfig.URL = config.GetString("url")
	client, err := marathon.NewClient(marathonConfig)
	if err != nil {
		return nil, err
	}
	a := MarathonInventory{log: log, Config: config, Client: client}
	return &a, nil
}

func (m *MarathonInventory) Total() (int, error) {
	app, err := m.GetApplication()
	if err != nil {
		return 0, err
	}
	return *app.Instances, nil
}

func (m *MarathonInventory) Increase() error {
	return m.Scale(+1)
}

func (m *MarathonInventory) Decrease() error {
	return m.Scale(-1)
}

func (m *MarathonInventory) Scale(amount int) error {
	// Check inventory status before trying to scale anything
	var e error
	app, err := m.GetApplication()
	if err != nil {
		return err
	}
	currentTotal, err := m.Total()
	if err != nil {
		return err
	}
	if m.Config.IsSet("minimum_instances") && currentTotal+amount < m.Config.GetInt("minimum_instances") {
		return errors.New("Won't scale below the minimum instances specified in config")
	}
	if m.Config.IsSet("maximum_instances") && currentTotal+amount > m.Config.GetInt("maximum_instances") {
		return errors.New("Won't scale above the maximum instances specified in config")
	}
	status, err := m.Status()
	if err != nil {
		return err
	}
	switch status {
	case UPDATING:
		e = errors.New("Won't scale application while another action is in progress")
	case FAILED:
		e = errors.New("Won't scale application while something seems to be in a failed state")
	case OK:
		if _, err := m.Client.ScaleApplicationInstances(app.ID, currentTotal+amount, false); err != nil {
			return err
		}
	default:
		e = errors.New("Unknown status")
	}
	if e == nil {
		m.log.Infof("Scaling %v by %v", app.ID, amount)
		m.lastModified = time.Now()
	}
	return e
}

func (m *MarathonInventory) Status() (Status, error) {
	app, err := m.GetApplication()
	if err != nil {
		return FAILED, err
	}
	if len(app.DeploymentIDs()) > 0 {
		return UPDATING, nil
	}
	if time.Now().Before(m.lastModified.Add(m.Config.GetDuration("settle_down_period"))) {
		m.log.Debugln("Still within settle down period")
		return UPDATING, nil
	}
	return OK, nil
}

func (m *MarathonInventory) GetApplication() (*marathon.Application, error) {
	name := m.Config.GetString("app")
	app, err := m.Client.ApplicationBy(name, &marathon.GetAppOpts{})
	if err != nil || app == nil {
		return nil, err
	}
	return app, nil
}
