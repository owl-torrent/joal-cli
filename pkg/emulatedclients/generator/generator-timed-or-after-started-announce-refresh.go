package generator

import (
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/algorithm"
	"github.com/pkg/errors"
	"time"
)

type TimedOrAfterStartedAnnounceRefreshGenerator struct {
	value          *string       `yaml:"-"`
	RefreshEvery   time.Duration `yaml:"refreshEvery"`
	nextGeneration time.Time     `yaml:"-"`
}

func (g *TimedOrAfterStartedAnnounceRefreshGenerator) Get(algorithm algorithm.IAlgorithm, infoHash torrent.InfoHash, event tracker.AnnounceEvent) string {
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

func (g *TimedOrAfterStartedAnnounceRefreshGenerator) AfterPropertiesSet() error {
	g.nextGeneration = time.Now()
	if g.RefreshEvery.Milliseconds() == 0 {
		return errors.New("'RefreshEvery' property can not be empty in TimedRefreshGenerator")
	}
	return nil
}
