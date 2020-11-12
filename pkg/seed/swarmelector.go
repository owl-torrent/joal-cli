package seed

import (
	"github.com/anthonyraymond/joal-cli/pkg/announcer"
	"github.com/anthonyraymond/joal-cli/pkg/bandwidth"
	"net/url"
	"sync"
	"time"
)

type peersStats struct {
	seeders     int32
	leechers    int32
	trackerHost string
	expiresAt   time.Time
}

func (s peersStats) GetSeeders() int32 {
	return s.seeders
}
func (s peersStats) GetLeechers() int32 {
	return s.leechers
}

type swarmUpdateRequest struct {
	trackerUrl url.URL
	interval   time.Duration
	seeders    int32
	leechers   int32
}

func errorSwarmUpdateRequest(trackerUrl url.URL) swarmUpdateRequest {
	return swarmUpdateRequest{
		trackerUrl: trackerUrl,
		interval:   1800 * time.Second,
		seeders:    0,
		leechers:   0,
	}
}
func successSwarmUpdateRequest(trackerUrl url.URL, response announcer.AnnounceResponse) swarmUpdateRequest {
	return swarmUpdateRequest{
		trackerUrl: trackerUrl,
		interval:   response.Interval * time.Second,
		seeders:    response.Seeders,
		leechers:   response.Leechers,
	}
}

func (s swarmUpdateRequest) toPeersStats() *peersStats {
	return &peersStats{
		seeders:     s.seeders,
		leechers:    s.leechers,
		trackerHost: s.trackerUrl.Host,
		expiresAt:   time.Now().Add(s.interval * 2),
	}
}

// SwarmElector act as a swarm, but it actually keep tracker seders in memory and elect the more representative in the list.
type swarmElector struct {
	elected    *peersStats
	hasChanged bool
	swarm      map[string]*peersStats
	lock       *sync.RWMutex
}

func newSwarmElector() *swarmElector {
	return &swarmElector{
		elected:    nil,
		hasChanged: false,
		swarm:      make(map[string]*peersStats),
		lock:       &sync.RWMutex{},
	}
}

func (s swarmElector) CurrentSwarm() bandwidth.ISwarm {
	s.lock.Lock()
	sw := peersStats{
		seeders:  0,
		leechers: 0,
	}
	if s.elected != nil {
		sw.seeders = s.elected.seeders
		sw.leechers = s.elected.leechers
	}
	s.lock.Unlock()

	return sw
}

func (s *swarmElector) GetSeeders() int32 {
	var v int32 = 0

	s.lock.RLock()
	if s.elected != nil {
		v = s.elected.seeders
	}
	s.lock.RUnlock()
	return v
}
func (s *swarmElector) GetLeechers() int32 {
	var v int32 = 0

	s.lock.RLock()
	if s.elected != nil {
		v = s.elected.leechers
	}
	s.lock.RUnlock()
	return v
}

// Update the warm with the new stats and return true if the swarm has changed, false if the swarm remains the same
func (s *swarmElector) UpdateSwarm(update swarmUpdateRequest) bool {
	newPeersStats := update.toPeersStats()
	s.lock.Lock()
	defer s.lock.Unlock()
	s.swarm[newPeersStats.trackerHost] = newPeersStats

	newElected := findBestAndEvictExpired(s.swarm)
	// has not changed
	if newElected != nil && s.elected != nil && s.elected.trackerHost == newElected.trackerHost && s.elected.leechers == newElected.leechers && s.elected.seeders == newElected.seeders {
		return false
	}
	s.elected = newElected
	return true
}

func findBestAndEvictExpired(swarm map[string]*peersStats) *peersStats {
	var max *peersStats
	for _, stats := range swarm {
		if stats.expiresAt.Before(time.Now()) {
			delete(swarm, stats.trackerHost)
			continue
		}
		if max == nil {
			max = stats
			continue
		}
		if max.seeders < stats.seeders {
			max = stats
		}
	}
	return max
}
