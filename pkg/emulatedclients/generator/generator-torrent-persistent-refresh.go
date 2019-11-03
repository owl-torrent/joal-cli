package generator

import (
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/algorithm"
	"github.com/anthonyraymond/joal-cli/pkg/utils"
	"sync"
	"time"
)

type TorrentPersistentGenerator struct {
	lock                sync.RWMutex                                  `yaml:"-"`
	entries             map[torrent.InfoHash]*utils.AccessAwareString `yaml:"-"`
	counterSinceCleanup int                                           `yaml:"-"`
	evictAfter          time.Duration                                 `yaml:"-"`
}

func (g *TorrentPersistentGenerator) Get(algorithm algorithm.IAlgorithm, infoHash torrent.InfoHash, event tracker.AnnounceEvent) string {
	g.lock.RLock()
	g.counterSinceCleanup += 1
	val, ok := g.entries[infoHash]
	g.lock.RUnlock()
	if !ok {
		g.lock.Lock()
		val = utils.AccessAwareStringNew(algorithm.Generate())
		g.entries[infoHash] = val
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
			for key, accessAware := range g.entries {
				if accessAware.LastAccess() > g.evictAfter {
					delete(g.entries, key)
				}
			}
		}()
	}

	return val.Get()
}

func (g *TorrentPersistentGenerator) AfterPropertiesSet() error {
	g.lock = sync.RWMutex{}
	g.entries = make(map[torrent.InfoHash]*utils.AccessAwareString, 10)
	g.counterSinceCleanup = 0
	g.evictAfter = 3600 * time.Second

	return nil
}
