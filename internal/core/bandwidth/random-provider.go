package bandwidth

import (
	"github.com/anthonyraymond/joal-cli/internal/core"
	"github.com/anthonyraymond/joal-cli/internal/utils/randutils"
	"sync"
)

type iRandomSpeedProvider interface {
	ReplaceSpeedConfig(conf *core.SpeedProviderConfig)
	GetBytesPerSeconds() int64
	Refresh()
}

type randomSpeedProvider struct {
	MinimumBytesPerSeconds int64
	MaximumBytesPerSeconds int64
	value                  int64
	lock                   *sync.RWMutex
}

func newRandomSpeedProvider(conf *core.SpeedProviderConfig) iRandomSpeedProvider {
	return &randomSpeedProvider{
		MinimumBytesPerSeconds: conf.MinimumBytesPerSeconds,
		MaximumBytesPerSeconds: conf.MaximumBytesPerSeconds,
		value:                  0,
		lock:                   &sync.RWMutex{},
	}
}

func (r *randomSpeedProvider) ReplaceSpeedConfig(conf *core.SpeedProviderConfig) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.MinimumBytesPerSeconds = conf.MinimumBytesPerSeconds
	r.MaximumBytesPerSeconds = conf.MaximumBytesPerSeconds
	r.value = randutils.Range(r.MinimumBytesPerSeconds, r.MaximumBytesPerSeconds)
}

func (r *randomSpeedProvider) GetBytesPerSeconds() int64 {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.value
}

func (r *randomSpeedProvider) Refresh() {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.value = randutils.Range(r.MinimumBytesPerSeconds, r.MaximumBytesPerSeconds)
}
