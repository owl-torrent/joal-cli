package algorithm

import (
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/key"
	"github.com/anthonyraymond/joal-cli/pkg/randutils"
)

type NumRangeAsHexAlgorithm struct {
	Min uint32 `yaml:"min"`
	Max uint32 `yaml:"max" validate:"min=1,gtefield=Min"`
}

func (a *NumRangeAsHexAlgorithm) Generate() key.Key {
	return key.Key(randutils.RangeUint32(a.Min, a.Max))
}

func (a *NumRangeAsHexAlgorithm) AfterPropertiesSet() error {
	return nil
}
