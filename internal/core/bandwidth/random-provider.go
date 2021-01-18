package bandwidth

import (
	"github.com/anthonyraymond/joal-cli/internal/core"
	"github.com/anthonyraymond/joal-cli/internal/utils/randutils"
)

type iRandomSpeedProvider interface {
	GetBytesPerSeconds() int64
	Refresh()
}

type randomSpeedProvider struct {
	MinimumBytesPerSeconds int64
	MaximumBytesPerSeconds int64
	value                  int64
}

func NewRandomSpeedProvider(conf *core.SpeedProviderConfig) iRandomSpeedProvider {
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
