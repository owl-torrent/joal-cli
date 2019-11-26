package generator

import (
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/key"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/key/algorithm"
	"github.com/pkg/errors"
	"time"
)

var generatorImplementations = map[string]func() iKeyGenerator{
	"NEVER_REFRESH":  func() iKeyGenerator { return &NeverRefreshGenerator{} },
	"ALWAYS_REFRESH": func() iKeyGenerator { return &AlwaysRefreshGenerator{} },
	"TIMED_REFRESH":  func() iKeyGenerator { return &TimedRefreshGenerator{} },
	"TIMED_OR_AFTER_STARTED_ANNOUNCE_REFRESH": func() iKeyGenerator { return &TimedOrAfterStartedAnnounceRefreshGenerator{} },
	"TORRENT_PERSISTENT_REFRESH":              func() iKeyGenerator { return &TorrentPersistentGenerator{} },
	"TORRENT_VOLATILE_REFRESH":                func() iKeyGenerator { return &TorrentVolatileGenerator{} },
}

type iKeyGenerator interface {
	Get(algorithm algorithm.IKeyAlgorithm, infoHash torrent.InfoHash, event tracker.AnnounceEvent) key.Key
	AfterPropertiesSet() error
}
type KeyGenerator struct {
	impl      iKeyGenerator           `yaml:",inline"`
	Algorithm algorithm.IKeyAlgorithm `yaml:"algorithm"`
}

func (a *KeyGenerator) UnmarshalYAML(unmarshal func(interface{}) error) error {
	unmarshalStruct := &struct {
		Name      string               `yaml:"type"`
		Algorithm *algorithm.KeyAlgorithm `yaml:"algorithm"`
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
		return errors.New(fmt.Sprintf("keyGenerator type '%s' does not exists. Possible values are: %v", unmarshalStruct.Name, allTypes))
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

func (a *KeyGenerator) Get(infoHash torrent.InfoHash, event tracker.AnnounceEvent) key.Key {
	return a.impl.Get(a.Algorithm, infoHash, event)
}

func (a *KeyGenerator) AfterPropertiesSet() error {
	if a.Algorithm == nil {
		return errors.New("NeverRefreshGenerator can not have a nil algorithm")
	}
	err := a.Algorithm.AfterPropertiesSet()
	if err != nil {
		return errors.Wrapf(err, "Failed to validate generator algorithm")
	}
	return a.impl.AfterPropertiesSet()
}


type AccessAwareKey struct {
	lastAccessed time.Time
	val          key.Key
}

func AccessAwareKeyNew(k key.Key) *AccessAwareKey {
	return &AccessAwareKey{
		lastAccessed: time.Now(),
		val:          k,
	}
}
func AccessAwareKeyNewSince(k key.Key, lastAccessed time.Time) *AccessAwareKey {
	return &AccessAwareKey{
		lastAccessed: lastAccessed,
		val:          k,
	}
}

func (s *AccessAwareKey) Get() key.Key {
	s.lastAccessed = time.Now()
	return s.val
}
func (s *AccessAwareKey) LastAccess() time.Duration {
	return time.Now().Sub(s.lastAccessed)
}