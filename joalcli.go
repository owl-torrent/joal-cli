package main

import (
	"context"
	"github.com/anthonyraymond/joal-cli/pkg/seedmanager"
	"github.com/anthonyraymond/joal-cli/pkg/seedmanager/config"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

// torrent base library : https://github.com/anacrolix/torrent
// especially for bencode and tracker subpackages

func init() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(os.Stdout)
	logrus.Info("init log to prevent race exception in logrus")
	logrus.Info("")
}

func main() {

	conf, err := config.ConfigManagerNew("C:/Users/raymo/Desktop/joal3/config.json")
	if err != nil {
		panic(err)
	}
	joal, err := seedmanager.JoalNew("C:/Users/raymo/Desktop/joal3", conf)
	if err != nil {
		panic(err)
	}

	err = joal.Start()
	if err != nil {
		panic(err)
	}

	timer := time.NewTimer(10 * time.Second)
	<-timer.C
	joal.Stop(nonCancellableTimeoutContext(5 * time.Second))
}

func nonCancellableTimeoutContext(duration time.Duration) context.Context {
	ctx, _ := context.WithTimeout(context.Background(), duration)
	return ctx
}
