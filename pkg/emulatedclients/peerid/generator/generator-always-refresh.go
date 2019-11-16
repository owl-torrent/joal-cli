package generator

import (
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/peerid"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/peerid/algorithm"
)

type AlwaysRefreshGenerator struct {
}

func (g *AlwaysRefreshGenerator) Get(algorithm algorithm.IPeerIdAlgorithm, infoHash torrent.InfoHash, event tracker.AnnounceEvent) peerid.PeerId {
	return algorithm.Generate()
}

func (g *AlwaysRefreshGenerator) AfterPropertiesSet() error {
	return nil
}
