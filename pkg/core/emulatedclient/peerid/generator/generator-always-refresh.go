package generator

import (
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/core/emulatedclient/peerid"
	"github.com/anthonyraymond/joal-cli/pkg/core/emulatedclient/peerid/algorithm"
)

type AlwaysRefreshGenerator struct {
}

func (g *AlwaysRefreshGenerator) get(algorithm algorithm.IPeerIdAlgorithm, infoHash torrent.InfoHash, event tracker.AnnounceEvent) peerid.PeerId {
	return algorithm.Generate()
}

func (g *AlwaysRefreshGenerator) afterPropertiesSet() error {
	return nil
}
