package algorithm

import (
	"crypto/rand"
	"fmt"
	"github.com/anthonyraymond/joal-cli/internal/old/core/emulatedclient/peerid"
	"io"
)

type PoolWithChecksumAlgorithm struct {
	randomSource   io.Reader `yaml:"-"`
	Prefix         string    `yaml:"prefix" validate:"required,lt=19"`
	CharactersPool string    `yaml:"charactersPool" validate:"required"`
}

func (a *PoolWithChecksumAlgorithm) Generate() peerid.PeerId {
	suffixLength := peerid.Length - len(a.Prefix)
	randomBytes := make([]byte, suffixLength-1)
	_, err := io.ReadFull(a.randomSource, randomBytes)
	if err != nil {
		panic(fmt.Sprintf("Failed to read random bytes: %+v", err))
	}

	buf := make([]byte, suffixLength)
	total := 0

	for i := 0; i < suffixLength-1; i++ {
		val := randomBytes[i]
		val = val % byte(len(a.CharactersPool))
		total = total + int(val)
		buf[i] = (a.CharactersPool)[val]
	}
	val := 0
	if total%len(a.CharactersPool) != 0 {
		val = len(a.CharactersPool) - (total % len(a.CharactersPool))
	}
	buf[suffixLength-1] = (a.CharactersPool)[val]
	var pid peerid.PeerId
	copy(pid[0:len(a.Prefix)], a.Prefix)
	copy(pid[len(a.Prefix):], buf)
	return pid
}

func (a *PoolWithChecksumAlgorithm) AfterPropertiesSet() error {
	a.randomSource = rand.Reader
	return nil
}
