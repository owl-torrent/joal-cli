package generator

import (
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/internal/core/emulatedclient/peerid"
	"github.com/anthonyraymond/joal-cli/internal/core/emulatedclient/peerid/algorithm"
	"sync"
)

type TorrentVolatileGenerator struct {
	lock                sync.RWMutex                            `yaml:"-"`
	entries             map[torrent.InfoHash]*AccessAwarePeerId `yaml:"-"`
	counterSinceCleanup int                                     `yaml:"-"`
}

func (g *TorrentVolatileGenerator) get(algorithm algorithm.IPeerIdAlgorithm, infoHash torrent.InfoHash, event tracker.AnnounceEvent) peerid.PeerId {
	g.lock.RLock()
	g.counterSinceCleanup += 1
	val, ok := g.entries[infoHash]
	g.lock.RUnlock()
	if !ok || val.IsExpired() {
		g.lock.Lock()
		val = accessAwarePeerIdNew(algorithm.Generate())
		g.entries[infoHash] = val
		g.lock.Unlock()
	}

	if event == tracker.Stopped {
		g.lock.Lock()
		delete(g.entries, infoHash)
		g.lock.Unlock()
	}

	// Once in a while clean the map
	if g.counterSinceCleanup > 100 {
		g.lock.Lock()
		if g.counterSinceCleanup > 100 {
			g.counterSinceCleanup = 0
			evictOldEntries(g.entries)
		}
		g.lock.Unlock()
	}

	return val.Get()
}

func (g *TorrentVolatileGenerator) afterPropertiesSet() error {
	g.lock = sync.RWMutex{}
	g.entries = make(map[torrent.InfoHash]*AccessAwarePeerId, 10)
	g.counterSinceCleanup = 0

	return nil
}
