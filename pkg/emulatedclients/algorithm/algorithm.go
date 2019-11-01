package algorithm

import (
	"fmt"
	"github.com/pkg/errors"
)

var algorithmImplementations = map[string]func() IAlgorithm{
	"HASH":                             func() IAlgorithm { return &HashAlgorithm{} },
	"REGEX":                            func() IAlgorithm { return &RegexPatternAlgorithm{} },
	"CHAR_POOL_WITH_CHECKSUM":          func() IAlgorithm { return &PoolWithChecksumAlgorithm{} },
	"NUM_RANGE_ENCODED_AS_HEXADECIMAL": func() IAlgorithm { return &NumRangeAsHexadecimalAlgorithm{} },
}

type IAlgorithm interface {
	Generate() string
	AfterPropertiesSet() error
}

type algorithm struct {
	Impl IAlgorithm `yaml:",inline"`
}

func (a *algorithm) UnmarshalYAML(unmarshal func(interface{}) error) error {
	algorithmType := &struct {
		Name string `yaml:"type"`
	}{}
	err := unmarshal(&algorithmType)
	if err != nil {
		return err
	}

	// if the algorithm is known create new empty instance of it
	implFactory, exist := algorithmImplementations[algorithmType.Name]
	if !exist {
		allTypes := make([]string, len(algorithmImplementations))
		i := 0
		for key := range algorithmImplementations {
			allTypes[i] = key
			i++
		}
		return errors.New(fmt.Sprintf("algorithm type '%s' does not exists. Possible values are: %v", algorithmType.Name, allTypes))
	}

	algorithm := implFactory()
	err = unmarshal(algorithm)
	if err != nil {
		return err
	}
	a.Impl = algorithm
	return nil
}

func (a *algorithm) Generate() string {
	return a.Impl.Generate()
}

func (a *algorithm) AfterPropertiesSet() error {
	return a.Impl.AfterPropertiesSet()
}
