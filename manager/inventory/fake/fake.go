package fake_inventory

import (
	"github.com/Sirupsen/logrus"
	"github.com/notonthehighstreet/autoscaler/manager/inventory"
	"github.com/spf13/viper"
)

type FakeInventory struct {
	config *viper.Viper
	log    *logrus.Entry
	total  int
}

func (f *FakeInventory) Total() (int, error) {
	return f.total, nil
}

func (f *FakeInventory) Increase() error {
	f.total += 1
	f.log.Infof("Fake inventory contains %v resources", f.total)
	return nil
}

func (f *FakeInventory) Decrease() error {
	f.total -= 1
	f.log.Infof("Fake inventory contains %v resources", f.total)
	return nil
}

func (f *FakeInventory) Status() (inventory.Status, error) {
	return inventory.OK, nil
}

func New(config *viper.Viper, log *logrus.Entry) inventory.Inventory {
	return &FakeInventory{config: config, log: log, total: 10}
}
