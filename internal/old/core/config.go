package core

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type JoalConfig struct {
	TorrentsDir         string
	ArchivedTorrentsDir string
	ClientsDir          string
	RuntimeConfig       *RuntimeConfig
}

func (c *JoalConfig) ListClientFiles() ([]string, error) {
	var clients []string

	err := filepath.Walk(c.ClientsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error at file '%s': %w", path, err)
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".yml" {
			return nil
		}
		clients = append(clients, filepath.Base(path))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error while walking though dir '%s': %w", c.ClientsDir, err)
	}

	return clients, nil
}

type RuntimeConfig struct {
	BandwidthConfig *BandwidthConfig `yaml:"bandwidth"`
	Client          string           `yaml:"client"`
}

// Return a new RuntimeConfig with the default values filled in
func (c RuntimeConfig) Default() *RuntimeConfig {
	return &RuntimeConfig{
		BandwidthConfig: BandwidthConfig{}.Default(),
		Client:          "qbittorrent-3.3.1.yml",
	}
}

type BandwidthConfig struct {
	Speed      *SpeedProviderConfig `yaml:"speed"`
	Dispatcher *DispatcherConfig    `yaml:"dispatcher"`
}

func (c BandwidthConfig) Default() *BandwidthConfig {
	return &BandwidthConfig{
		Speed:      SpeedProviderConfig{}.Default(),
		Dispatcher: DispatcherConfig{}.Default(),
	}
}

type DispatcherConfig struct {
	GlobalBandwidthRefreshInterval time.Duration `yaml:"globalBandwidthRefreshInterval"`
}

func (c DispatcherConfig) Default() *DispatcherConfig {
	return &DispatcherConfig{
		GlobalBandwidthRefreshInterval: 20 * time.Minute,
	}
}

type SpeedProviderConfig struct {
	MinimumBytesPerSeconds int64 `yaml:"min"`
	MaximumBytesPerSeconds int64 `yaml:"max"`
}

func (c SpeedProviderConfig) Default() *SpeedProviderConfig {
	return &SpeedProviderConfig{
		MinimumBytesPerSeconds: 5000,
		MaximumBytesPerSeconds: 15000,
	}
}
