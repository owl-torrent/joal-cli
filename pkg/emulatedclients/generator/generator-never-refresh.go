package generator

import (
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/algorithm"
)

type NeverRefreshGenerator struct {
	value *string `yaml:"-"`
}

func (g *NeverRefreshGenerator) Get(algorithm algorithm.IAlgorithm, infoHash torrent.InfoHash, event tracker.AnnounceEvent) string {
	if g.value == nil {
		val := algorithm.Generate()
		g.value = &val
	}
	return *g.value
}

func (g *NeverRefreshGenerator) AfterPropertiesSet() error {
	return nil
}
