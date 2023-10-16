package algorithm

import (
	"fmt"
	"github.com/anthonyraymond/joal-cli/internal/old/core/emulatedclient/peerid"
	"gopkg.in/yaml.v3"
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

func (a *PeerIdAlgorithm) UnmarshalYAML(value *yaml.Node) error {
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
		for key := range algorithmImplementations {
			allTypes[i] = key
			i++
		}
		return fmt.Errorf("algorithm type '%s' does not exists. Possible values are: %v", algorithmType.Name, allTypes)
	}

	algorithm := implFactory()
	err = value.Decode(algorithm)
	if err != nil {
		return err
	}
	a.IPeerIdAlgorithm = algorithm
	return nil
}
