package config

import (
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

type SeedConfig struct {
	MinUploadRate              int64  `json:"minUploadRate" yaml:"maxUploadRate"`
	MaxUploadRate              int64  `json:"maxUploadRate" yaml:"maxUploadRate"`
	Client                     string `json:"clientFile" yaml:"clientFile"`
	RemoveTorrentWithZeroPeers bool   `json:"removeTorrentWithZeroPeers" yaml:"removeTorrentWithZeroPeers"`
}

type Manager struct {
	configPath string
	seedConfig *SeedConfig
}

func ConfigManagerNew(configFilePath string) (*Manager, error) {
	if !filepath.IsAbs(configFilePath) {
		return nil, errors.New("config file path must be an absolute path")
	}
	return &Manager{
		configPath: configFilePath,
		seedConfig: nil,
	}, nil
}

func (c *Manager) Save(config SeedConfig) error {
	jsonStr, err := json.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "failed to marshall config")
	}

	err = ioutil.WriteFile(c.configPath, jsonStr, 0644)
	if err != nil {
		return errors.Wrap(err, "failed to write config")
	}
	return nil
}

func (c *Manager) Load() (SeedConfig, error) {
	jsonFile, err := os.Open(c.configPath)
	if err != nil {
		return SeedConfig{}, errors.Wrapf(err, "failed to open file '%s'", c.configPath)
	}
	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return SeedConfig{}, errors.Wrapf(err, "failed to read file '%s'", c.configPath)
	}

	var seedConfig SeedConfig

	err = json.Unmarshal(bytes, &seedConfig)
	if err != nil {
		return SeedConfig{}, errors.Wrapf(err, "failed to parse json config '%s'", string(bytes))
	}

	c.seedConfig = &seedConfig

	return seedConfig, nil
}

func (c *Manager) Get() (SeedConfig, error) {
	if c.seedConfig == nil {
		_, err := c.Load()
		if err != nil {
			return SeedConfig{}, errors.Wrap(err, "failed to load config")
		}
	}

	return *c.seedConfig, nil
}
