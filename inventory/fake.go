package inventory

import (
	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
)

type Fake struct {
	config *viper.Viper
	log    *logrus.Entry
	total  int
}

func (f *Fake) Total() (int, error) {
	return f.total, nil
}

func (f *Fake) Increase() error {
	f.total += 1
	f.log.Infof("Fake inventory contains %v resources", f.total)
	return nil
}

func (f *Fake) Decrease() error {
	f.total -= 1
	f.log.Infof("Fake inventory contains %v resources", f.total)
	return nil
}

func (f *Fake) Status() (Status, error) {
	return OK, nil
}

func NewFake(config *viper.Viper, log *logrus.Entry) (Inventory, error) {
	return &Fake{config: config, log: log, total: 10}, nil
}
