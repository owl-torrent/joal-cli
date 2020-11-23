package config

import (
	"github.com/anthonyraymond/joal-cli/pkg/bandwidth"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"testing"
	"time"
)

func TestRuntimeConfig_ShouldUnmarshal(t *testing.T) {
	yamlStr := `
client: coco.client
bandwidth:
  dispatcher:
    globalBandwidthRefreshInterval: 300h
    incrementSeedInterval: 1h
  speed:
    min: 10
    max: 100
`

	c := &RuntimeConfig{}
	err := yaml.Unmarshal([]byte(yamlStr), c)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, &RuntimeConfig{
		Client: "coco.client",
		BandwidthConfig: &bandwidth.BandwidthConfig{
			Dispatcher: &bandwidth.DispatcherConfig{
				GlobalBandwidthRefreshInterval:           300 * time.Hour,
				IntervalBetweenEachTorrentsSeedIncrement: 1 * time.Hour,
			},
			Speed: &bandwidth.SpeedProviderConfig{
				MinimumBytesPerSeconds: 10,
				MaximumBytesPerSeconds: 100,
			},
		},
	}, c)
}

func TestRuntimeConfig_ShouldUnmarshalAndReplaceDefault(t *testing.T) {
	yamlStr := `
client: coco.client
bandwidth:
  dispatcher:
    globalBandwidthRefreshInterval: 300h
    incrementSeedInterval: 1h
  speed:
    min: 10
    max: 100
`

	c := &RuntimeConfig{}
	err := yaml.Unmarshal([]byte(yamlStr), c)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, &RuntimeConfig{
		Client: "coco.client",
		BandwidthConfig: &bandwidth.BandwidthConfig{
			Dispatcher: &bandwidth.DispatcherConfig{
				GlobalBandwidthRefreshInterval:           300 * time.Hour,
				IntervalBetweenEachTorrentsSeedIncrement: 1 * time.Hour,
			},
			Speed: &bandwidth.SpeedProviderConfig{
				MinimumBytesPerSeconds: 10,
				MaximumBytesPerSeconds: 100,
			},
		},
	}, c)
}
