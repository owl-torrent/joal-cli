package main

import (
	"context"
	"github.com/anthonyraymond/joal-cli/pkg/logs"
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
	defer logs.GetLogger().Sync()
	conf, err := config.ConfigManagerNew(os.Args[1])
	if err != nil {
		panic(err)
	}
	joal, err := seedmanager.JoalNew(os.Args[2], conf)
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
