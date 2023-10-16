package generator

import (
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/internal/old/core/emulatedclient/key"
	"github.com/anthonyraymond/joal-cli/internal/old/core/emulatedclient/key/algorithm"
	"gopkg.in/yaml.v3"
	"time"
)

var generatorImplementations = map[string]func() IKeyGenerator{
	"NEVER_REFRESH":  func() IKeyGenerator { return &NeverRefreshGenerator{} },
	"ALWAYS_REFRESH": func() IKeyGenerator { return &AlwaysRefreshGenerator{} },
	"TIMED_REFRESH":  func() IKeyGenerator { return &TimedRefreshGenerator{} },
	"TIMED_OR_AFTER_STARTED_ANNOUNCE_REFRESH": func() IKeyGenerator { return &TimedOrAfterStartedAnnounceRefreshGenerator{} },
	"TORRENT_PERSISTENT_REFRESH":              func() IKeyGenerator { return &TorrentPersistentGenerator{} },
	"TORRENT_VOLATILE_REFRESH":                func() IKeyGenerator { return &TorrentVolatileGenerator{} },
}

type IKeyGenerator interface {
	get(algorithm algorithm.IKeyAlgorithm, infoHash torrent.InfoHash, event tracker.AnnounceEvent) key.Key
	afterPropertiesSet() error
}
type KeyGenerator struct {
	IKeyGenerator `yaml:",inline" validate:"required"`
	Algorithm     algorithm.IKeyAlgorithm `yaml:"algorithm" validate:"required"`
}

func (a *KeyGenerator) UnmarshalYAML(value *yaml.Node) error {
	unmarshalStruct := &struct {
		Name      string                  `yaml:"type"`
		Algorithm *algorithm.KeyAlgorithm `yaml:"algorithm"`
	}{}
	if a.Algorithm != nil {
		unmarshalStruct.Algorithm = a.Algorithm.(*algorithm.KeyAlgorithm)
	}
	err := value.Decode(&unmarshalStruct)
	if err != nil {
		return err
	}

	implFactory, exist := generatorImplementations[unmarshalStruct.Name]
	if !exist {
		allTypes := make([]string, len(generatorImplementations))
		i := 0
		for generatorType := range generatorImplementations {
			allTypes[i] = generatorType
			i++
		}
		return fmt.Errorf("keyGenerator type '%s' does not exists. Possible values are: %v", unmarshalStruct.Name, allTypes)
	}

	// if the generator is known create new empty instance of it
	generator := implFactory()
	err = value.Decode(generator)
	if err != nil {
		return err
	}
	a.IKeyGenerator = generator
	a.Algorithm = unmarshalStruct.Algorithm
	return nil
}

func (a *KeyGenerator) Get(infoHash torrent.InfoHash, event tracker.AnnounceEvent) key.Key {
	return a.IKeyGenerator.get(a.Algorithm, infoHash, event)
}

func (a *KeyGenerator) AfterPropertiesSet() error {
	err := a.Algorithm.AfterPropertiesSet()
	if err != nil {
		return fmt.Errorf("failed to validate generator algorithm: %w", err)
	}
	return a.IKeyGenerator.afterPropertiesSet()
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

func (s *AccessAwareKey) Get() key.Key {
	s.lastAccessed = time.Now()
	return s.val
}

// time elapsed since last access
func (s *AccessAwareKey) IsExpired() bool {
	return time.Since(s.lastAccessed) > 3*time.Hour
}

func evictOldEntries(entries map[torrent.InfoHash]*AccessAwareKey) {
	for k, accessAware := range entries {
		if accessAware.IsExpired() {
			delete(entries, k)
		}
	}
}
