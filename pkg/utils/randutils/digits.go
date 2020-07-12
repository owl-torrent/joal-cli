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

// Return a random number between min and max
func RangeUint32(minInclusive uint32, maxInclusive uint32) uint32 {
	if minInclusive == maxInclusive {
		return minInclusive
	}

	return uint32(Range(int64(minInclusive), int64(maxInclusive)))
}
