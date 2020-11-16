package bandwidth

import "time"

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
	GlobalBandwidthRefreshInterval           time.Duration `yaml:"globalBandwidthRefreshInterval"`
	IntervalBetweenEachTorrentsSeedIncrement time.Duration `yaml:"incrementSeedInterval"`
}

func (c DispatcherConfig) Default() *DispatcherConfig {
	return &DispatcherConfig{
		GlobalBandwidthRefreshInterval:           20 * time.Minute,
		IntervalBetweenEachTorrentsSeedIncrement: 5 * time.Second,
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
