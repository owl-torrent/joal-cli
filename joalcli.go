package main

import (
	"context"
	"fmt"
	"github.com/anthonyraymond/joal-cli/internal/old/core"
	"github.com/anthonyraymond/joal-cli/internal/old/core/logs"
	"github.com/anthonyraymond/joal-cli/internal/old/core/manager2"
	"github.com/anthonyraymond/joal-cli/internal/old/plugins"
	"github.com/anthonyraymond/joal-cli/internal/old/plugins/types"
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

const configRootFolder = `D:\temp\trash\joaltest`

func main() {
	defer func() { _ = logs.GetLogger().Sync() }()

	configLocation := getConfigRootFolder()

	appConfig, err := BootstrapApp(configLocation)
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

	coreConfigLoader, err := core.Bootstrap(filepath.Join(configLocation, "core"), httpClient)
	if err != nil {
		panic(err)
	}

	manager, _ := manager2.Run(coreConfigLoader)
	bridge := types.NewCoreBridge(coreConfigLoader, manager)
	pluginManager := plugins.NewPluginManager(configLocation, bridge)
	pluginManager.BootstrapPlugins(httpClient)
	pluginManager.StartPlugins()

	manager.StartSeeding()
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
	pluginManager.ShutdownPlugins(ctxPlugin)
}

func getDefaultConfigFolder() (string, error) {
	// Windows => %AppData%/joal
	// Mac     => $HOME/Library/Application Support/joal
	// Linux   => $XDG_CONFIG_HOME/joal or $HOME/.config/joal
	dir, err := os.UserConfigDir()
	return filepath.Join(dir, "joal"), err
}

func getConfigRootFolder() string {
	configLocation := configRootFolder // TODO: remove me and get from os.args
	var err error
	if strings.TrimSpace(configLocation) == "" {
		configLocation, err = getDefaultConfigFolder()
		if err != nil {
			panic(fmt.Errorf("failed to resolve default config folder: %w", err))
		}
	}
	configLocation, err = filepath.Abs(configLocation)
	if err != nil {
		panic(fmt.Errorf("failed to transform '%s' to an absolute path: %w", configLocation, err))
	}
	return configLocation
}
