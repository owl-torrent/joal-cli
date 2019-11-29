package algorithm

import (
	"fmt"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/peerid"
	"github.com/pkg/errors"
)

var algorithmImplementations = map[string]func() IPeerIdAlgorithm{
	"REGEX":                   func() IPeerIdAlgorithm { return &RegexPatternAlgorithm{} },
	"CHAR_POOL_WITH_CHECKSUM": func() IPeerIdAlgorithm { return &PoolWithChecksumAlgorithm{} },
}

type IPeerIdAlgorithm interface {
	Generate() peerid.PeerId
	AfterPropertiesSet() error
}

type PeerIdAlgorithm struct {
	IPeerIdAlgorithm `yaml:",inline" validate:"required"`
}

func (a *PeerIdAlgorithm) UnmarshalYAML(unmarshal func(interface{}) error) error {
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
	a.IPeerIdAlgorithm = algorithm
	return nil
}
