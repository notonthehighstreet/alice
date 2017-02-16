package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/evalphobia/logrus_fluent"
	"github.com/heirko/go-contrib/logrusHelper"
	"github.com/johntdyer/slackrus"
	"github.com/notonthehighstreet/alice"
	conf "github.com/spf13/viper"
	"sync"
	"time"
)

func init() {
	// Register plugins at load time
	alice.RegisterInventory("aws", alice.NewAWSInventory)
	alice.RegisterInventory("fake", alice.NewFakeInventory)
	alice.RegisterInventory("marathon", alice.NewMarathonInventory)
	alice.RegisterMonitor("fake", alice.NewFakeMonitor)
	alice.RegisterMonitor("mesos", alice.NewMesosMonitor)
	alice.RegisterMonitor("datadog", alice.NewDatadogMonitor)
	alice.RegisterStrategy("ratio", alice.NewRatioStrategy)
	alice.RegisterStrategy("threshold", alice.NewThresholdStrategy)

	// Setup config
	conf.AddConfigPath("./config")
	if err := conf.ReadInConfig(); err != nil {
		logrus.Panicf("Fatal error config file: %s \n", err)
	}
	conf.SetDefault("interval", 2*time.Minute)
	conf.SetDefault("logging.level", "info")
}

func main() {
	log := initLogger()
	var managers []alice.Manager
	for name := range conf.GetStringMap("managers") {
		if mgr, err := alice.New(conf.Sub("managers."+name), log.WithField("manager", name)); err != nil {
			log.Fatalf("Error initializing manager: %s", err.Error())
		} else {
			managers = append(managers, mgr)
		}
	}
	for {
		var wg sync.WaitGroup
		wg.Add(len(managers))
		for _, man := range managers {
			go func(m alice.Manager) {
				defer wg.Done()
				m.Run()
			}(man)
		}
		wg.Wait()
		time.Sleep(conf.GetDuration("interval"))
	}
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
		fluentConf.SetDefault("tag", "service.alice")

		hook, err := logrus_fluent.New(fluentConf.GetString("host"), fluentConf.GetInt("port"))
		if err != nil {
			logrus.Panic(err)
		}
		hook.SetTag(fluentConf.GetString("tag"))
		logrus.AddHook(hook)
	}
	if loggingConf.IsSet("slack") {
		slackConf := loggingConf.Sub("slack")
		slackConf.SetDefault("username", "alice")
		slackConf.SetDefault("emoji", ":robot_face:")
		slackConf.SetDefault("channel", "#slack-testing")
		if !slackConf.IsSet("hook_url") {
			logrus.Fatalln("Must provide hook_url for Slack.")
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
