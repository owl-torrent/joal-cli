package generator

import (
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/key"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/key/algorithm"
	"sync"
	"time"
)

type TorrentVolatileGenerator struct {
	lock                sync.RWMutex                               `yaml:"-"`
	entries             map[torrent.InfoHash]*AccessAwareKey `yaml:"-"`
	counterSinceCleanup int                                        `yaml:"-"`
	evictAfter          time.Duration                              `yaml:"-"`
}

func (g *TorrentVolatileGenerator) Get(algorithm algorithm.IKeyAlgorithm, infoHash torrent.InfoHash, event tracker.AnnounceEvent) key.Key {
	g.lock.RLock()
	g.counterSinceCleanup += 1
	val, ok := g.entries[infoHash]
	g.lock.RUnlock()
	if !ok {
		g.lock.Lock()
		val = AccessAwareKeyNew(algorithm.Generate())
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
		g.counterSinceCleanup = 0
		g.lock.Unlock()
		go func() {
			g.lock.Lock()
			defer g.lock.Unlock()
			for k, accessAware := range g.entries {
				if accessAware.LastAccess() > g.evictAfter {
					delete(g.entries, k)
				}
			}
		}()
	}

	return val.Get()
}

func (g *TorrentVolatileGenerator) AfterPropertiesSet() error {
	g.lock = sync.RWMutex{}
	g.entries = make(map[torrent.InfoHash]*AccessAwareKey, 10)
	g.counterSinceCleanup = 0
	g.evictAfter = 3600 * time.Second

	return nil
}