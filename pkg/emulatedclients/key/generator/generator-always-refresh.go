package generator

import (
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/key"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/key/algorithm"
)

type AlwaysRefreshGenerator struct {
}

func (g *AlwaysRefreshGenerator) Get(algorithm algorithm.IKeyAlgorithm, infoHash torrent.InfoHash, event tracker.AnnounceEvent) key.Key {
	return algorithm.Generate()
}

func (g *AlwaysRefreshGenerator) AfterPropertiesSet() error {
	return nil
}
