package config

import (
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
		BandwidthConfig: &BandwidthConfig{
			Dispatcher: &DispatcherConfig{
				GlobalBandwidthRefreshInterval:           300 * time.Hour,
				IntervalBetweenEachTorrentsSeedIncrement: 1 * time.Hour,
			},
			Speed: &SpeedProviderConfig{
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
		BandwidthConfig: &BandwidthConfig{
			Dispatcher: &DispatcherConfig{
				GlobalBandwidthRefreshInterval:           300 * time.Hour,
				IntervalBetweenEachTorrentsSeedIncrement: 1 * time.Hour,
			},
			Speed: &SpeedProviderConfig{
				MinimumBytesPerSeconds: 10,
				MaximumBytesPerSeconds: 100,
			},
		},
	}, c)
}

func TestBandwidthConfig_ShouldUnmarshal(t *testing.T) {
	yamlStr := `
dispatcher:
  globalBandwidthRefreshInterval: 300h
  incrementSeedInterval: 1h
speed:
  min: 10
  max: 100
`

	c := &BandwidthConfig{}
	err := yaml.Unmarshal([]byte(yamlStr), c)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, &BandwidthConfig{
		Dispatcher: &DispatcherConfig{
			GlobalBandwidthRefreshInterval:           300 * time.Hour,
			IntervalBetweenEachTorrentsSeedIncrement: 1 * time.Hour,
		},
		Speed: &SpeedProviderConfig{
			MinimumBytesPerSeconds: 10,
			MaximumBytesPerSeconds: 100,
		},
	}, c)
}

func TestBandwidthConfig_ShouldUnmarshalAndReplaceDefault(t *testing.T) {
	yamlStr := `
speed:
  min: 10
  max: 100
`

	c := BandwidthConfig{}.Default()
	err := yaml.Unmarshal([]byte(yamlStr), c)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, DispatcherConfig{}.Default(), c.Dispatcher)
	assert.Equal(t, &SpeedProviderConfig{
		MinimumBytesPerSeconds: 10,
		MaximumBytesPerSeconds: 100,
	}, c.Speed)
}

func TestDispatcherConfig_ShouldUnmarshal(t *testing.T) {
	yamlStr := `
globalBandwidthRefreshInterval: 300h
incrementSeedInterval: 1h
`

	c := &DispatcherConfig{}
	err := yaml.Unmarshal([]byte(yamlStr), c)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, &DispatcherConfig{
		GlobalBandwidthRefreshInterval:           300 * time.Hour,
		IntervalBetweenEachTorrentsSeedIncrement: 1 * time.Hour,
	}, c)
}

func TestDispatcherConfig_ShouldUnmarshalAndReplaceDefault(t *testing.T) {
	yamlStr := `
globalBandwidthRefreshInterval: 300h
`

	c := DispatcherConfig{}.Default()
	err := yaml.Unmarshal([]byte(yamlStr), c)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, &DispatcherConfig{
		GlobalBandwidthRefreshInterval:           300 * time.Hour,
		IntervalBetweenEachTorrentsSeedIncrement: DispatcherConfig{}.Default().IntervalBetweenEachTorrentsSeedIncrement,
	}, c)
}

func TestSpeedProviderConfig_ShouldUnmarshal(t *testing.T) {
	yamlStr := `
min: 25
max: 10000
`

	c := &SpeedProviderConfig{}
	err := yaml.Unmarshal([]byte(yamlStr), c)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, &SpeedProviderConfig{
		MinimumBytesPerSeconds: 25,
		MaximumBytesPerSeconds: 10000,
	}, c)
}

func TestSpeedProviderConfig_ShouldUnmarshalAndReplaceDefault(t *testing.T) {
	yamlStr := `
max: 10000
`

	c := SpeedProviderConfig{}.Default()
	err := yaml.Unmarshal([]byte(yamlStr), c)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, &SpeedProviderConfig{
		MinimumBytesPerSeconds: SpeedProviderConfig{}.Default().MinimumBytesPerSeconds,
		MaximumBytesPerSeconds: 10000,
	}, c)
}
