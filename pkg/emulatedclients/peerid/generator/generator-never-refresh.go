package generator

import (
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/peerid"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/peerid/algorithm"
)

type NeverRefreshGenerator struct {
	value *peerid.PeerId `yaml:"-"`
}

func (g *NeverRefreshGenerator) Get(algorithm algorithm.IPeerIdAlgorithm, infoHash torrent.InfoHash, event tracker.AnnounceEvent) peerid.PeerId {
	if g.value == nil {
		val := algorithm.Generate()
		g.value = &val
	}
	return *g.value
}

func (g *NeverRefreshGenerator) AfterPropertiesSet() error {
	return nil
}
