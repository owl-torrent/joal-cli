package main

import (
	"context"
	"github.com/anthonyraymond/joal-cli/pkg/logs"
	"github.com/anthonyraymond/joal-cli/pkg/seedmanager"
	"github.com/anthonyraymond/joal-cli/pkg/seedmanager/config"
	"os"
	"time"
)

// torrent base library : https://github.com/anacrolix/torrent
// especially for bencode and tracker subpackages

func main() {
	defer func() { _ = logs.GetLogger().Sync() }()
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	joal.Stop(ctx)
	cancel()
}
