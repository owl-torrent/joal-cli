package algorithm

import (
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/key"
	"github.com/anthonyraymond/joal-cli/pkg/randutils"
	"github.com/pkg/errors"
)

type NumRangeAsHexAlgorithm struct {
	Min uint32 `yaml:"min"`
	Max uint32 `yaml:"max"`
}

func (a *NumRangeAsHexAlgorithm) Generate() key.Key {
	return key.Key(randutils.RangeUint32(a.Min, a.Max))
}

func (a *NumRangeAsHexAlgorithm) AfterPropertiesSet() error {
	if a.Min > a.Max {
		return errors.New("'max' must be greater or equal to 'min' in NumRangeAsHexAlgorithm")
	}

	return nil
}
