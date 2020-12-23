package generator

import (
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/core/emulatedclient/key"
	"github.com/anthonyraymond/joal-cli/pkg/core/emulatedclient/key/algorithm"
	"time"
)

type TimedRefreshGenerator struct {
	value          *key.Key      `yaml:"-"`
	RefreshEvery   time.Duration `yaml:"refreshEvery" validate:"required"`
	nextGeneration time.Time     `yaml:"-"`
}

func (g *TimedRefreshGenerator) get(algorithm algorithm.IKeyAlgorithm, infoHash torrent.InfoHash, event tracker.AnnounceEvent) key.Key {
	if g.shouldRegenerate() {
		val := algorithm.Generate()
		g.value = &val
		g.nextGeneration = time.Now().Add(g.RefreshEvery)
	}
	return *g.value
}

func (g *TimedRefreshGenerator) shouldRegenerate() bool {
	if g.value == nil {
		return true
	}
	if g.nextGeneration.Before(time.Now()) {
		return true
	}
	return false
}

func (g *TimedRefreshGenerator) afterPropertiesSet() error {
	g.nextGeneration = time.Now()
	return nil
}
