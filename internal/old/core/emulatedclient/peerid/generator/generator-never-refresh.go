package generator

import (
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/internal/old/core/emulatedclient/peerid"
	"github.com/anthonyraymond/joal-cli/internal/old/core/emulatedclient/peerid/algorithm"
)

type NeverRefreshGenerator struct {
	value *peerid.PeerId `yaml:"-"`
}

func (g *NeverRefreshGenerator) get(algorithm algorithm.IPeerIdAlgorithm, infoHash torrent.InfoHash, event tracker.AnnounceEvent) peerid.PeerId {
	if g.value == nil {
		val := algorithm.Generate()
		g.value = &val
	}
	return *g.value
}

func (g *NeverRefreshGenerator) afterPropertiesSet() error {
	return nil
}
