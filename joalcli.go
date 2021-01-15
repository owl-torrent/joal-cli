package main

import (
	"context"
	"github.com/anthonyraymond/joal-cli/pkg/core/config"
	"github.com/anthonyraymond/joal-cli/pkg/core/logs"
	"github.com/anthonyraymond/joal-cli/pkg/core/seedmanager"
	"github.com/anthonyraymond/joal-cli/pkg/plugins"
	"github.com/anthonyraymond/joal-cli/pkg/plugins/web"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net"
	"net/http"
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

func getConfigRootFolder() string {
	configLocation := configRootFolder // TODO: remove me and get from os.args
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
	return configLocation
}

func bootstrapAndGetConfig(configLocation string) (*AppConfig, error) {
	err := BootstrapApp(configLocation)
	if err != nil {
		panic(err)
	}
	return ParseConfigOverDefault(configLocation)
}

func main() {
	defer logs.GetLogger().Sync()

	configLocation := getConfigRootFolder()

	appConfig, err := bootstrapAndGetConfig(configLocation)
	if err != nil {
		panic(err)
	}

	err = logs.ReplaceLogger(appConfig.Log)
	if err != nil {
		panic(err)
	}

	// Create an Http Client (almost similar to http.DefaultClient, but we apply our proxy configuration)
	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy: appConfig.Proxy.Proxy(),
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
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

	for _, p := range enabledPlugins {
		if err := p.Initialize(filepath.Join(configLocation, "plugins"), httpClient); err != nil {
			logs.GetLogger().Error("failed to initialize plugin, this plugin will stay off for the rest of the execution", zap.String("plugin", p.Name()), zap.Error(err))
		}
	}

	coreConfigLoader, err := config.NewJoalConfigLoader(filepath.Join(configLocation, "core"), httpClient)
	if err != nil {
		panic(err)
	}
	manager := seedmanager.NewTorrentManager(coreConfigLoader)

	coreBridge := plugins.NewCoreBridge(manager, coreConfigLoader)
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
