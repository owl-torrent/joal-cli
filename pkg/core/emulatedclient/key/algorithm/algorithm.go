package algorithm

import (
	"fmt"
	"github.com/anthonyraymond/joal-cli/pkg/core/emulatedclient/key"
	"gopkg.in/yaml.v3"
)

var algorithmImplementations = map[string]func() IKeyAlgorithm{
	"NUM_RANGE_ENCODED_AS_HEXADECIMAL": func() IKeyAlgorithm { return &NumRangeAsHexAlgorithm{} },
}

type IKeyAlgorithm interface {
	Generate() key.Key
	AfterPropertiesSet() error
}

type KeyAlgorithm struct {
	IKeyAlgorithm `yaml:",inline" validate:"required"`
}

func (a *KeyAlgorithm) UnmarshalYAML(value *yaml.Node) error {
	algorithmType := &struct {
		Name string `yaml:"type"`
	}{}
	err := value.Decode(&algorithmType)
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
		return fmt.Errorf("algorithm type '%s' does not exists. Possible values are: %v", algorithmType.Name, allTypes)
	}

	algorithm := implFactory()
	err = value.Decode(algorithm)
	if err != nil {
		return err
	}
	a.IKeyAlgorithm = algorithm
	return nil
}
