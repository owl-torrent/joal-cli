package generator

import (
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/algorithm"
	"github.com/pkg/errors"
	"time"
)

type TimedRefreshGenerator struct {
	value          *string       `yaml:"-"`
	RefreshEvery   time.Duration `yaml:"refreshEvery"`
	nextGeneration time.Time     `yaml:"-"`
}

func (g *TimedRefreshGenerator) Get(algorithm algorithm.IAlgorithm, infoHash torrent.InfoHash, event tracker.AnnounceEvent) string {
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

func (g *TimedRefreshGenerator) AfterPropertiesSet() error {
	g.nextGeneration = time.Now()
	if g.RefreshEvery.Milliseconds() == 0 {
		return errors.New("'RefreshEvery' property can not be empty in TimedRefreshGenerator")
	}
	return nil
}
