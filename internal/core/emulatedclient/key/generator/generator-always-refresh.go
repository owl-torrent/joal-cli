package generator

import (
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/internal/core/emulatedclient/key"
	"github.com/anthonyraymond/joal-cli/internal/core/emulatedclient/key/algorithm"
)

type AlwaysRefreshGenerator struct {
}

func (g *AlwaysRefreshGenerator) get(algorithm algorithm.IKeyAlgorithm, infoHash torrent.InfoHash, event tracker.AnnounceEvent) key.Key {
	return algorithm.Generate()
}

func (g *AlwaysRefreshGenerator) afterPropertiesSet() error {
	return nil
}
