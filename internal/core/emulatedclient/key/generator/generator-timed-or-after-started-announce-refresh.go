package generator

import (
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/internal/core/emulatedclient/key"
	"github.com/anthonyraymond/joal-cli/internal/core/emulatedclient/key/algorithm"
	"time"
)

type TimedOrAfterStartedAnnounceRefreshGenerator struct {
	value          *key.Key      `yaml:"-"`
	RefreshEvery   time.Duration `yaml:"refreshEvery" validate:"required"`
	nextGeneration time.Time     `yaml:"-"`
}

func (g *TimedOrAfterStartedAnnounceRefreshGenerator) get(algorithm algorithm.IKeyAlgorithm, infoHash torrent.InfoHash, event tracker.AnnounceEvent) key.Key {
	if g.shouldRegenerate(event) {
		val := algorithm.Generate()
		g.value = &val
		g.nextGeneration = time.Now().Add(g.RefreshEvery)
	}
	return *g.value
}

func (g *TimedOrAfterStartedAnnounceRefreshGenerator) shouldRegenerate(event tracker.AnnounceEvent) bool {
	if g.value == nil {
		return true
	}
	if event == tracker.Started {
		return true
	}
	if g.nextGeneration.Before(time.Now()) {
		return true
	}
	return false
}

func (g *TimedOrAfterStartedAnnounceRefreshGenerator) afterPropertiesSet() error {
	g.nextGeneration = time.Now()
	return nil
}
