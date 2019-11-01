package generator

import (
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/algorithm"
)

type AlwaysRefreshGenerator struct {
}

func (g *AlwaysRefreshGenerator) Get(algorithm algorithm.IAlgorithm, infoHash torrent.InfoHash, event tracker.AnnounceEvent) string {
	return algorithm.Generate()
}

func (g *AlwaysRefreshGenerator) AfterPropertiesSet() error {
	return nil
}
