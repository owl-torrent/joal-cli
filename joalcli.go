package main

import (
	"context"
	"github.com/anthonyraymond/joal-cli/pkg/core/logs"
	"github.com/anthonyraymond/joal-cli/pkg/core/seedmanager"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// torrent base library : https://github.com/anacrolix/torrent
// especially for bencode and tracker subpackages

func main() {
	logs.SetLevel(zap.DebugLevel)
	manager, err := seedmanager.NewTorrentManager(`D:\temp\trash\joaltest`)
	if err != nil {
		panic(err)
	}

	err = manager.Seed()
	if err != nil {
		panic(err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	manager.StopSeeding(ctx)
}
