package main

import (
	"fmt"
	"github.com/anthonyraymond/joal-cli/internal/old/common/configloader"
	"github.com/anthonyraymond/joal-cli/internal/old/core/logs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var (
	configFileFromRoot = func(rootConfigDir string) string {
		return filepath.Join(rootConfigDir, "config.yml")
	}
)

func BootstrapApp(configDir string) (*AppConfig, error) {
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create folder '%s': %w", configDir, err)
	}

	// Create the configuration file if missing
	f, err := os.OpenFile(configFileFromRoot(configDir), os.O_CREATE, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create '%s' file: %w", configFileFromRoot(configDir), err)
	}
	_ = f.Close()

	return readConfig(configDir)
}

func readConfig(configDir string) (*AppConfig, error) {
	conf := AppConfig{}.Default()
	err := configloader.ParseIntoDefault(configFileFromRoot(configDir), conf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

type AppConfig struct {
	Log   *logs.LogConfig `yaml:"log"`
	Proxy ProxyConf       `yaml:"proxy"`
}

func (ac AppConfig) Default() *AppConfig {
	return &AppConfig{
		Log:   logs.LogConfig{}.Default(),
		Proxy: ProxyConf{}.Default(),
	}
}

type ProxyConf struct {
	Url string `yaml:"url"`
}

func (pc ProxyConf) Default() ProxyConf {
	return ProxyConf{
		Url: "",
	}
}

func (pc *ProxyConf) Proxy() func(*http.Request) (*url.URL, error) {
	if strings.TrimSpace(pc.Url) == "" {
		return nil
	}

	u, err := url.Parse(pc.Url)
	return func(request *http.Request) (*url.URL, error) {
		return u, err
	}
}
