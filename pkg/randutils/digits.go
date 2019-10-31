package randutils

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	mrand "math/rand"
)

func NewCryptoSeededSource() mrand.Source {
	var seed int64
	_ = binary.Read(cryptorand.Reader, binary.BigEndian, &seed)
	return mrand.NewSource(seed)
}

var globalRand *mrand.Rand = mrand.New(NewCryptoSeededSource())

// Return a random number between min and max
func Range(minInclusive int64, maxInclusive int64) int64 {
	if minInclusive == maxInclusive {
		return minInclusive
	}
	if minInclusive >= 0 {
		return globalRand.Int63n(maxInclusive-minInclusive) + minInclusive
	}
	// at this point the min inclusive is assured to be less than 0
	return globalRand.Int63n(-minInclusive+maxInclusive) + minInclusive
}
