package alice

import (
	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
)

// FakeInventory doesn't control a real inventory, instead it stores a count of dummy objects
type FakeInventory struct {
	config *viper.Viper
	log    *logrus.Entry
	total  int
}

// Total returns the current total number of resources
func (f *FakeInventory) Total() (int, error) {
	return f.total, nil
}

// Increase (scale up) the number of resources in the inventory
func (f *FakeInventory) Increase() error {
	f.total++
	f.log.Infof("Fake inventory contains %v resources", f.total)
	return nil
}

// Decrease (scale down) the number of resources in the inventory
func (f *FakeInventory) Decrease() error {
	f.total--
	f.log.Infof("Fake inventory contains %v resources", f.total)
	return nil
}

// Status returns OK if the inventory is ready to be scaled, UPDATING if an update is in progress, or FAILED
func (f *FakeInventory) Status() (Status, error) {
	return OK, nil
}

// NewFakeInventory creates a new Inventory
func NewFakeInventory(config *viper.Viper, log *logrus.Entry) (Inventory, error) {
	return &FakeInventory{config: config, log: log, total: 10}, nil
}
