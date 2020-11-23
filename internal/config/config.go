package config

import (
	"github.com/anthonyraymond/joal-cli/pkg/bandwidth"
)

type JoalConfig struct {
	torrentsDir         string
	archivedTorrentsDir string
	clientsDir          string
	runtimeConfig       *RuntimeConfig
}

type RuntimeConfig struct {
	BandwidthConfig *bandwidth.BandwidthConfig `yaml:"bandwidth"`
	Client          string                     `yaml:"client"`
}

// Return a new RuntimeConfig with the default values filled in
func (c RuntimeConfig) Default() *RuntimeConfig {
	return &RuntimeConfig{
		BandwidthConfig: bandwidth.BandwidthConfig{}.Default(),
		Client:          "",
	}
}
