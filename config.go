package main

import (
	"github.com/anthonyraymond/joal-cli/internal/common/configloader"
	"github.com/anthonyraymond/joal-cli/internal/core/logs"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var (
	configFileName = func(rootConfigDir string) string {
		return filepath.Join(rootConfigDir, "config.yml")
	}
)

func BootstrapApp(configDir string) error {
	// Create the configuration file if missing
	f, err := os.OpenFile(configFileName(configDir), os.O_CREATE, 0755)
	if err != nil {
		return errors.Wrapf(err, "failed to create '%s' file", configFileName(configDir))
	}
	defer func() { _ = f.Close() }()
	return nil
}

func ParseConfigOverDefault(configDir string) (*AppConfig, error) {
	conf, err := configloader.NewConfigLoader(configFileName(configDir), func() interface{} {
		return AppConfig{}.Default()
	}).ParseOverDefault()
	if err != nil {
		return nil, err
	}
	c, ok := conf.(*AppConfig)
	if !ok {
		panic("fuck")
	}

	return c, nil
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
