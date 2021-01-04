package main

import (
	"context"
	"github.com/anthonyraymond/joal-cli/pkg/core/logs"
	"github.com/anthonyraymond/joal-cli/pkg/core/seedmanager"
	"github.com/anthonyraymond/joal-cli/pkg/plugins"
	"github.com/anthonyraymond/joal-cli/pkg/plugins/web"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

// torrent base library : https://github.com/anacrolix/torrent
// especially for bencode and tracker subpackages

func main() {
	plugs := []plugins.IJoalPlugin{
		&web.Plugin{},
	}

	logs.SetLevel(zap.DebugLevel)

	for _, p := range plugs {
		if err := p.Initialize(filepath.Join(`D:\temp\trash\joaltest`, p.SubFolder())); err != nil {
			logs.GetLogger().Error("failed to initialize plugin, this plugin will stay off for the rest of the execution", zap.String("plugin", p.SubFolder()), zap.Error(err))
		}
	}

	manager, err := seedmanager.NewTorrentManager(`D:\temp\trash\joaltest\core`)
	if err != nil {
		panic(err)
	}

	coreBridge := plugins.NewCoreBridge(manager)
	for _, p := range plugs {
		if p.Enabled() {
			p.AfterCoreLoaded(coreBridge)
		}
	}

	err = manager.StartSeeding()
	if err != nil {
		panic(err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	manager.StopSeeding(ctx)

	ctxPlugin, cancelPlugin := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelPlugin()
	for _, p := range plugs {
		if p.Enabled() {
			p.Shutdown(ctxPlugin)
		}
	}
}
