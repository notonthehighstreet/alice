package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/heirko/go-contrib/logrusHelper"
	"github.com/notonthehighstreet/autoscaler/manager"
	conf "github.com/spf13/viper"
	"time"
)

func main() {
	configure()
	var c = logrusHelper.UnmarshalConfiguration(conf.Sub("logging")) // Unmarshal configuration from Viper
	logrusHelper.SetConfig(logrus.StandardLogger(), c)               // for e.g. apply it to logrus default instance
	managers := make(map[string]manager.Manager)

	for name := range conf.GetStringMap("managers") {
		managers[name] = manager.New(conf.Sub("managers."+name), logrus.WithField("manager", name))
	}
	interval := conf.GetDuration("interval")
	for range time.NewTicker(interval).C {
		for _, manager := range managers {
			manager.Run()
		}
	}
}

func configure() {
	conf.AddConfigPath(".")
	err := conf.ReadInConfig()
	if err != nil {
		logrus.Panicf("Fatal error config file: %s \n", err)
	}

	conf.SetDefault("interval", 2*time.Minute)
	conf.SetDefault("logging.level", "info")
}
