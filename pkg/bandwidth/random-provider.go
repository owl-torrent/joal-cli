package bandwidth

import "github.com/anthonyraymond/joal-cli/pkg/utils/randutils"

type IRandomSpeedProvider interface {
	GetBytesPerSeconds() int64
	Refresh()
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

type randomSpeedProvider struct {
	MinimumBytesPerSeconds int64
	MaximumBytesPerSeconds int64
	value                  int64
}

func newRandomSpeedProvider(conf *SpeedProviderConfig) IRandomSpeedProvider {
	return &randomSpeedProvider{
		MinimumBytesPerSeconds: conf.MinimumBytesPerSeconds,
		MaximumBytesPerSeconds: conf.MaximumBytesPerSeconds,
		value:                  0,
	}
}

func (r *randomSpeedProvider) GetBytesPerSeconds() int64 {
	return r.value
}

func (r *randomSpeedProvider) Refresh() {
	r.value = randutils.Range(r.MinimumBytesPerSeconds, r.MaximumBytesPerSeconds)
}
