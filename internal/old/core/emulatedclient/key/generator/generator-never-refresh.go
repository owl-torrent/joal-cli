package generator

import (
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/internal/old/core/emulatedclient/key"
	"github.com/anthonyraymond/joal-cli/internal/old/core/emulatedclient/key/algorithm"
)

type NeverRefreshGenerator struct {
	value *key.Key `yaml:"-"`
}

func (g *NeverRefreshGenerator) get(algorithm algorithm.IKeyAlgorithm, infoHash torrent.InfoHash, event tracker.AnnounceEvent) key.Key {
	if g.value == nil {
		val := algorithm.Generate()
		g.value = &val
	}
	return *g.value
}

func (g *NeverRefreshGenerator) afterPropertiesSet() error {
	return nil
}
