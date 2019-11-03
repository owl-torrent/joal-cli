package generator

import (
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/algorithm"
	"github.com/pkg/errors"
)

var generatorImplementations = map[string]func() IGenerator{
	"NEVER_REFRESH":  func() IGenerator { return &NeverRefreshGenerator{} },
	"ALWAYS_REFRESH": func() IGenerator { return &AlwaysRefreshGenerator{} },
	"TIMED_REFRESH":  func() IGenerator { return &TimedRefreshGenerator{} },
	"TIMED_OR_AFTER_STARTED_ANNOUNCE_REFRESH": func() IGenerator { return &TimedOrAfterStartedAnnounceRefreshGenerator{} },
	"TORRENT_PERSISTENT_REFRESH":              func() IGenerator { return &TorrentPersistentGenerator{} },
}

type IGenerator interface {
	Get(algorithm algorithm.IAlgorithm, infoHash torrent.InfoHash, event tracker.AnnounceEvent) string
	AfterPropertiesSet() error
}
type generator struct {
	Impl      IGenerator           `yaml:",inline"`
	Algorithm algorithm.IAlgorithm `yaml:"algorithm"`
}

func (a *generator) UnmarshalYAML(unmarshal func(interface{}) error) error {
	generatorType := &struct {
		Name string `yaml:"type"`
	}{}
	err := unmarshal(&generatorType)
	if err != nil {
		return err
	}

	// if the generator is known create new empty instance of it
	implFactory, exist := generatorImplementations[generatorType.Name]
	if !exist {
		allTypes := make([]string, len(generatorImplementations))
		i := 0
		for key := range generatorImplementations {
			allTypes[i] = key
			i++
		}
		return errors.New(fmt.Sprintf("generator type '%s' does not exists. Possible values are: %v", generatorType.Name, allTypes))
	}

	generator := implFactory()
	err = unmarshal(generator)
	if err != nil {
		return err
	}
	a.Impl = generator
	return nil
}

func (a *generator) Get(infoHash torrent.InfoHash, event tracker.AnnounceEvent) string {
	return a.Impl.Get(a.Algorithm, infoHash, event)
}

func (a *generator) AfterPropertiesSet() error {
	if a.Algorithm == nil {
		return errors.New("NeverRefreshGenerator can not have a nil algorithm")
	}
	err := a.Algorithm.AfterPropertiesSet()
	if err != nil {
		return errors.Wrapf(err, "Failed to validate generator algorithm")
	}
	return a.Impl.AfterPropertiesSet()
}
