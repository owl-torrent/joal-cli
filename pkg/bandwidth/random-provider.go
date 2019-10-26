package bandwidth

import (
	"crypto/rand"
	"math/big"
)

type IRandomSpeedProvider interface {
	GetBytesPerSeconds() int64
	Refresh()
}
type RandomSpeedProvider struct {
	minimumBytesPerSeconds int64
	maximumBytesPerSeconds int64
	value                  int64
}

func (r *RandomSpeedProvider) GetBytesPerSeconds() int64 {
	return r.value
}
func (r *RandomSpeedProvider) Refresh() {
	upperBound := big.NewInt(0).Sub(big.NewInt(r.maximumBytesPerSeconds), big.NewInt(r.minimumBytesPerSeconds))
	n, err := rand.Int(rand.Reader, upperBound)
	if err != nil {
		r.value = 0
		return
	}
	r.value = n.Int64() + r.minimumBytesPerSeconds
}
