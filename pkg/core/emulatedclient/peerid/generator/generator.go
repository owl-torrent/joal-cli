package generator

import (
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/core/emulatedclient/peerid"
	"github.com/anthonyraymond/joal-cli/pkg/core/emulatedclient/peerid/algorithm"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"time"
)

var generatorImplementations = map[string]func() IPeerIdGenerator{
	"NEVER_REFRESH":  func() IPeerIdGenerator { return &NeverRefreshGenerator{} },
	"ALWAYS_REFRESH": func() IPeerIdGenerator { return &AlwaysRefreshGenerator{} },
	"TIMED_REFRESH":  func() IPeerIdGenerator { return &TimedRefreshGenerator{} },
	"TIMED_OR_AFTER_STARTED_ANNOUNCE_REFRESH": func() IPeerIdGenerator { return &TimedOrAfterStartedAnnounceRefreshGenerator{} },
	"TORRENT_PERSISTENT_REFRESH":              func() IPeerIdGenerator { return &TorrentPersistentGenerator{} },
	"TORRENT_VOLATILE_REFRESH":                func() IPeerIdGenerator { return &TorrentVolatileGenerator{} },
}

type IPeerIdGenerator interface {
	get(algorithm algorithm.IPeerIdAlgorithm, infoHash torrent.InfoHash, event tracker.AnnounceEvent) peerid.PeerId
	afterPropertiesSet() error
}
type PeerIdGenerator struct {
	IPeerIdGenerator `yaml:",inline" validate:"required"`
	Algorithm        algorithm.IPeerIdAlgorithm `yaml:"algorithm" validate:"required"`
}

func (a *PeerIdGenerator) UnmarshalYAML(value *yaml.Node) error {
	unmarshalStruct := &struct {
		Name      string                     `yaml:"type"`
		Algorithm *algorithm.PeerIdAlgorithm `yaml:"algorithm"`
	}{}
	err := value.Decode(&unmarshalStruct)
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
		return fmt.Errorf("peerIdGenerator type '%s' does not exists. Possible values are: %v", unmarshalStruct.Name, allTypes)
	}

	generator := implFactory()
	err = value.Decode(generator)
	if err != nil {
		return err
	}
	a.IPeerIdGenerator = generator
	a.Algorithm = unmarshalStruct.Algorithm
	return nil
}

func (a *PeerIdGenerator) Get(infoHash torrent.InfoHash, event tracker.AnnounceEvent) peerid.PeerId {
	return a.IPeerIdGenerator.get(a.Algorithm, infoHash, event)
}

func (a *PeerIdGenerator) AfterPropertiesSet() error {
	err := a.Algorithm.AfterPropertiesSet()
	if err != nil {
		return errors.Wrapf(err, "Failed to validate generator algorithm")
	}
	return a.IPeerIdGenerator.afterPropertiesSet()
}

type AccessAwarePeerId struct {
	lastAccessed time.Time
	val          peerid.PeerId
}

func accessAwarePeerIdNew(k peerid.PeerId) *AccessAwarePeerId {
	return &AccessAwarePeerId{
		lastAccessed: time.Now(),
		val:          k,
	}
}

func (s *AccessAwarePeerId) Get() peerid.PeerId {
	s.lastAccessed = time.Now()
	return s.val
}

func (s *AccessAwarePeerId) IsExpired() bool {
	return time.Since(s.lastAccessed) > 3*time.Hour
}

func evictOldEntries(entries map[torrent.InfoHash]*AccessAwarePeerId) {
	for key, accessAware := range entries {
		if accessAware.IsExpired() {
			delete(entries, key)
		}
	}
}
