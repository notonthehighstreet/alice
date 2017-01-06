package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/evalphobia/logrus_fluent"
	"github.com/heirko/go-contrib/logrusHelper"
	"github.com/johntdyer/slackrus"
	"github.com/notonthehighstreet/autoscaler/manager"
	conf "github.com/spf13/viper"
	"time"
)

func main() {
	configure()
	log := initLogger()
	managers := make(map[string]manager.Manager)

	for name := range conf.GetStringMap("managers") {
		managers[name] = manager.New(conf.Sub("managers."+name), log.WithField("manager", name))
	}
	interval := conf.GetDuration("interval")
	for range time.NewTicker(interval).C {
		for _, manager := range managers {
			manager.Run()
		}
	}
}

func configure() {
	conf.AddConfigPath("./config")
	err := conf.ReadInConfig()
	if err != nil {
		logrus.Panicf("Fatal error config file: %s \n", err)
	}

	conf.SetDefault("interval", 2*time.Minute)
	conf.SetDefault("logging.level", "info")
}

func initLogger() *logrus.Entry {
	loggingConf := conf.Sub("logging")
	var c = logrusHelper.UnmarshalConfiguration(loggingConf) // Unmarshal configuration from Viper
	logrusHelper.SetConfig(logrus.StandardLogger(), c)       // apply it to logrus default instance

	// We're on our own for fluentd
	if loggingConf.IsSet("fluentd") {
		fluentConf := loggingConf.Sub("fluentd")

		fluentConf.SetDefault("host", "172.17.42.1")
		fluentConf.SetDefault("port", 24224)
		fluentConf.SetDefault("tag", "service.autoscaler")

		hook, err := logrus_fluent.New(fluentConf.GetString("host"), fluentConf.GetInt("port"))
		if err != nil {
			logrus.Panic(err)
		}
		hook.SetTag(fluentConf.GetString("tag"))
		logrus.AddHook(hook)
	}
	if loggingConf.IsSet("slack") {
		slackConf := loggingConf.Sub("slack")
		slackConf.SetDefault("username", "autoscaler")
		slackConf.SetDefault("emoji", ":robot_face:")
		slackConf.SetDefault("channel", "#slack-testing")
		if !slackConf.IsSet("hook_url") {
			logrus.Fatalln("Must provide hook_url for slack.")
		}
		logrus.AddHook(&slackrus.SlackrusHook{
			HookURL:        slackConf.GetString("hook_url"),
			AcceptedLevels: slackrus.LevelThreshold(logrus.WarnLevel),
			Channel:        slackConf.GetString("channel"),
			IconEmoji:      slackConf.GetString("emoji"),
			Username:       slackConf.GetString("username"),
		})

	}
	fields := logrus.Fields{}
	if loggingConf.IsSet("custom_fields") {
		for k, v := range loggingConf.GetStringMapString("custom_fields") {
			fields[k] = v
		}
	}
	return logrus.StandardLogger().WithFields(fields)
}
