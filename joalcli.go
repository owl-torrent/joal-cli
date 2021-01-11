package main

import (
	"context"
	"github.com/anthonyraymond/joal-cli/pkg/core/logs"
	"github.com/anthonyraymond/joal-cli/pkg/core/seedmanager"
	"github.com/anthonyraymond/joal-cli/pkg/plugins"
	"github.com/anthonyraymond/joal-cli/pkg/plugins/web"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// torrent base library : https://github.com/anacrolix/torrent
// especially for bencode and tracker subpackages

var availablePlugins = []plugins.IJoalPlugin{
	&web.Plugin{},
}

const configRootFolder = `D:\temp\trash\joaltest`

func main() {
	configLocation := configRootFolder
	var err error
	if strings.TrimSpace(configLocation) == "" {
		configLocation, err = getDefaultConfigFolder()
		if err != nil {
			panic(errors.Wrap(err, "failed to resolve default config folder"))
		}
	}
	configLocation, err = filepath.Abs(configLocation)
	if err != nil {
		panic(errors.Wrapf(err, "failed to transform '%s' to an absolute path", configLocation))
	}


	log := logs.GetLogger()
	var enabledPlugins []plugins.IJoalPlugin


	log.Info("evaluate plugins list", zap.Any("available-plugins", pluginListToListOfName(availablePlugins)))
	for _, p := range availablePlugins {
		if p.ShouldEnable() {
			enabledPlugins = append(enabledPlugins, p)
		}
	}

	log.Info("plugins list has been initialized", zap.Any("enabled-plugins", pluginListToListOfName(enabledPlugins)))


	//TODO: remove me
	logs.SetLevel(zap.DebugLevel)

	for _, p := range enabledPlugins {
		if err := p.Initialize(filepath.Join(configLocation, "plugins")); err != nil {
			logs.GetLogger().Error("failed to initialize plugin, this plugin will stay off for the rest of the execution", zap.String("plugin", p.Name()), zap.Error(err))
		}
	}

	manager, err := seedmanager.NewTorrentManager(filepath.Join(configLocation, "core"))
	if err != nil {
		panic(err)
	}

	coreBridge := plugins.NewCoreBridge(manager)
	for _, p := range enabledPlugins {
		p.AfterCoreLoaded(coreBridge)
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
	for _, p := range enabledPlugins {
		p.Shutdown(ctxPlugin)
	}
}

func pluginListToListOfName(pluginList []plugins.IJoalPlugin) []string {
	var names []string
	for _, p := range pluginList {
		names = append(names, p.Name())
	}
	return names
}

func getDefaultConfigFolder() (string, error) {
	// Windows => %AppData%/joal
	// Mac     => $HOME/Library/Application Support/joal
	// Linux   => $XDG_CONFIG_HOME/joal or $HOME/.config/joal
	dir, err := os.UserConfigDir()
	return filepath.Join(dir, "joal"), err
}