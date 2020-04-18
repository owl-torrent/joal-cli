package bandwidth

//go:generate mockgen -destination=./random-provider_mock.go -self_package=github.com/anthonyraymond/joal-cli/pkg/randutils -package=bandwidth github.com/anthonyraymond/joal-cli/pkg/bandwidth IRandomSpeedProvider

import "github.com/anthonyraymond/joal-cli/pkg/randutils"

type IRandomSpeedProvider interface {
	GetBytesPerSeconds() int64
	Refresh()
}
type RandomSpeedProvider struct {
	MinimumBytesPerSeconds int64
	MaximumBytesPerSeconds int64
	value                  int64
}

func (r *RandomSpeedProvider) GetBytesPerSeconds() int64 {
	return r.value
}
func (r *RandomSpeedProvider) Refresh() {
	r.value = randutils.Range(r.MinimumBytesPerSeconds, r.MaximumBytesPerSeconds)
}
