package generator

import (
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/algorithm"
	"github.com/pkg/errors"
)

var generatorImplementations = map[string]func() iGenerator{
	"NEVER_REFRESH":  func() iGenerator { return &NeverRefreshGenerator{} },
	"ALWAYS_REFRESH": func() iGenerator { return &AlwaysRefreshGenerator{} },
	"TIMED_REFRESH":  func() iGenerator { return &TimedRefreshGenerator{} },
	"TIMED_OR_AFTER_STARTED_ANNOUNCE_REFRESH": func() iGenerator { return &TimedOrAfterStartedAnnounceRefreshGenerator{} },
	"TORRENT_PERSISTENT_REFRESH":              func() iGenerator { return &TorrentPersistentGenerator{} },
	"TORRENT_VOLATILE_REFRESH":                func() iGenerator { return &TorrentVolatileGenerator{} },
}

type iGenerator interface {
	Get(algorithm algorithm.IAlgorithm, infoHash torrent.InfoHash, event tracker.AnnounceEvent) string
	AfterPropertiesSet() error
}
type Generator struct {
	impl      iGenerator           `yaml:",inline"`
	Algorithm algorithm.IAlgorithm `yaml:"algorithm"`
}

func (a *Generator) UnmarshalYAML(unmarshal func(interface{}) error) error {
	unmarshalStruct := &struct {
		Name      string               `yaml:"type"`
		Algorithm *algorithm.Algorithm `yaml:"algorithm"`
	}{}
	err := unmarshal(&unmarshalStruct)
	if err != nil {
		return err
	}

	// if the generator is known create new empty instance of it
	implFactory, exist := generatorImplementations[unmarshalStruct.Name]
	if !exist {
		allTypes := make([]string, len(generatorImplementations))
		i := 0
		for key := range generatorImplementations {
			allTypes[i] = key
			i++
		}
		return errors.New(fmt.Sprintf("generator type '%s' does not exists. Possible values are: %v", unmarshalStruct.Name, allTypes))
	}

	generator := implFactory()
	err = unmarshal(generator)
	if err != nil {
		return err
	}
	a.impl = generator
	a.Algorithm = unmarshalStruct.Algorithm
	return nil
}

func (a *Generator) Get(infoHash torrent.InfoHash, event tracker.AnnounceEvent) string {
	return a.impl.Get(a.Algorithm, infoHash, event)
}

func (a *Generator) AfterPropertiesSet() error {
	if a.Algorithm == nil {
		return errors.New("NeverRefreshGenerator can not have a nil algorithm")
	}
	err := a.Algorithm.AfterPropertiesSet()
	if err != nil {
		return errors.Wrapf(err, "Failed to validate generator algorithm")
	}
	return a.impl.AfterPropertiesSet()
}
