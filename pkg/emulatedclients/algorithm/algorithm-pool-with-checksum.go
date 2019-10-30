package algorithm

import (
	"crypto/rand"
	"fmt"
	"github.com/pkg/errors"
	"io"
)

type PoolWithChecksumAlgorithm struct {
	RandomSource   io.Reader
	Prefix         *string `yaml:"prefix"`
	CharactersPool *string `yaml:"charactersPool"`
	Length         *int    `yaml:"length"`
}

func (a *PoolWithChecksumAlgorithm) Generate() string {
	suffixLength := *a.Length - len(*a.Prefix)
	randomBytes := make([]byte, suffixLength-1)
	_, err := io.ReadFull(a.RandomSource, randomBytes)
	if err != nil {
		panic(fmt.Sprintf("Failed to read random bytes: %+v", err))
	}

	buf := make([]byte, suffixLength)
	total := 0

	for i := 0; i < suffixLength-1; i++ {
		val := randomBytes[i]
		val = val % byte(len(*a.CharactersPool))
		total = total + int(val)
		buf[i] = (*a.CharactersPool)[val]
	}
	val := 0
	if total%len(*a.CharactersPool) != 0 {
		val = len(*a.CharactersPool) - (total % len(*a.CharactersPool))
	}
	buf[suffixLength-1] = (*a.CharactersPool)[val]
	return *a.Prefix + string(buf)
}

func (a *PoolWithChecksumAlgorithm) AfterPropertiesSet() error {
	a.RandomSource = rand.Reader
	if a.Prefix == nil {
		return errors.New("PoolWithChecksumAlgorithm can not have a null prefix")
	}
	if len(*a.Prefix) > 18 {
		return errors.Errorf("PoolWithChecksumAlgorithm prefix is too long '%s'", *a.Prefix)
	}
	if a.CharactersPool == nil {
		return errors.New("PoolWithChecksumAlgorithm can not have a null charactersPool")
	}
	if len(*a.CharactersPool) < 1 {
		return errors.Errorf("PoolWithChecksumAlgorithm charactersPool is too short '%s'", *a.CharactersPool)
	}
	if a.Length == nil {
		return errors.New("PoolWithChecksumAlgorithm can not have a null length")
	}
	if *a.Length <= len(*a.Prefix)+1 {
		return errors.Errorf("PoolWithChecksumAlgorithm length is too short '%s'", *a.CharactersPool)
	}

	return nil
}
