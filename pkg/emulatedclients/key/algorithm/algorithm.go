package algorithm

import (
	"fmt"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/key"
	"github.com/pkg/errors"
)

var algorithmImplementations = map[string]func() IKeyAlgorithm{
	"NUM_RANGE_ENCODED_AS_HEXADECIMAL": func() IKeyAlgorithm { return &NumRangeAsHexAlgorithm{} },
}

type IKeyAlgorithm interface {
	Generate() key.Key
	AfterPropertiesSet() error
}

type KeyAlgorithm struct {
	impl IKeyAlgorithm `yaml:",inline" validate:"required"`
}

func (a *KeyAlgorithm) UnmarshalYAML(unmarshal func(interface{}) error) error {
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
		for algKey := range algorithmImplementations {
			allTypes[i] = algKey
			i++
		}
		return errors.New(fmt.Sprintf("algorithm type '%s' does not exists. Possible values are: %v", algorithmType.Name, allTypes))
	}

	algorithm := implFactory()
	err = unmarshal(algorithm)
	if err != nil {
		return err
	}
	a.impl = algorithm
	return nil
}

func (a *KeyAlgorithm) Generate() key.Key {
	return a.impl.Generate()
}
func (a *KeyAlgorithm) AfterPropertiesSet() error {
	return a.impl.AfterPropertiesSet()
}
