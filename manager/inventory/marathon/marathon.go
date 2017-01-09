package marathon

import (
	"github.com/spf13/viper"
	"time"
	"github.com/notonthehighstreet/autoscaler/manager/inventory"
	"github.com/Sirupsen/logrus"
	"github.com/gambol99/go-marathon"
	"errors"
)

type MarathonClient interface {
	ApplicationBy(name string, opts *marathon.GetAppOpts) (*marathon.Application, error)
	ScaleApplicationInstances(name string, instances int, force bool) (*marathon.DeploymentID, error)
}

type MarathonInventory struct {
	log            *logrus.Entry
	Client 	       MarathonClient
	Config         *viper.Viper
	lastModified   time.Time
}

func New(config *viper.Viper, log *logrus.Entry) inventory.Inventory {
	requiredConfig := []string{"app", "url"}
	for _, item := range requiredConfig {
		if !config.IsSet(item) {
			log.Fatalf("Missing config: %v", item)
		}
	}
	config.SetDefault("settle_down_period", "0s")
	marathonConfig := marathon.NewDefaultConfig()
	marathonConfig.URL = config.GetString("url")
	client, err := marathon.NewClient(marathonConfig)
	if err != nil {
	    log.Fatalf("Failed to create a client for marathon, error: %s", err)
	}
	a := MarathonInventory{log: log, Config: config, Client: client}
	return &a
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
	switch m.Status() {
	case inventory.UPDATING:
		e = errors.New("Won't scale application while another action is in progress")
	case inventory.FAILED:
		e = errors.New("Won't scale application while something seems to be in a failed state")
	case inventory.OK:
		if _, err := m.Client.ScaleApplicationInstances(app.ID, currentTotal + amount, false); err != nil {
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

func (m *MarathonInventory) Status() inventory.Status {
	app, err := m.GetApplication()
	if err != nil {
		return inventory.FAILED
	}
	if len(app.DeploymentIDs()) > 0 {
		return inventory.UPDATING
	}
	if time.Now().Before(m.lastModified.Add(m.Config.GetDuration("settle_down_period"))) {
		m.log.Debugln("Still within settle down period")
		return inventory.UPDATING
	}
	return inventory.OK
}

func (m *MarathonInventory) GetApplication() (*marathon.Application, error) {
	name := m.Config.GetString("app")
	app, err := m.Client.ApplicationBy(name, &marathon.GetAppOpts{})
	if err != nil || app == nil {
		return nil, err
	}
	return app, nil
}