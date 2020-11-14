package seedmanager

import "github.com/anthonyraymond/joal-cli/pkg/bandwidth"

type RuntimeConfig struct {
	BandwidthConfig *bandwidth.Config `yaml:"bandwidth"`
	Client          string            `yaml:"client"`
}

// Return a new RuntimeConfig with the default values filled in
func (c RuntimeConfig) Default() *RuntimeConfig {
	return &RuntimeConfig{
		BandwidthConfig: bandwidth.Config{}.Default(),
		Client:          "",
	}
}
