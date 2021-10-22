package torrent2

import (
	"github.com/anthonyraymond/joal-cli/internal/core/announcer"
	"net/url"
	"sync"
	"time"
)

type Peers interface {
	Seeders() int32
	Leechers() int32
	AddPeer(update SwarmUpdateRequest)
	Reset()
}

type peer struct {
	seeders     int32
	leechers    int32
	trackerHost string
	expiresAt   time.Time
}

type peersElector struct {
	peers       map[string]peer
	electedPeer peer
	lock        *sync.Mutex
}

func newPeersElector() Peers {
	return &peersElector{
		peers:       make(map[string]peer),
		electedPeer: peer{},
		lock:        &sync.Mutex{},
	}
}

func (p *peersElector) Seeders() int32 {
	p.lock.Lock()
	v := p.electedPeer.seeders
	p.lock.Unlock()
	return v
}

func (p *peersElector) Leechers() int32 {
	p.lock.Lock()
	v := p.electedPeer.leechers
	p.lock.Unlock()
	return v
}

func (p *peersElector) AddPeer(update SwarmUpdateRequest) {
	newPeersStats := update.toPeersStats()

	p.lock.Lock()
	defer func() { p.lock.Unlock() }()

	p.peers[newPeersStats.trackerHost] = peer{seeders: newPeersStats.seeders, leechers: newPeersStats.leechers}
	p.proceedToPeerElection()
}

func (p *peersElector) Reset() {
	p.lock.Lock()

	p.peers = make(map[string]peer)
	p.electedPeer = peer{}

	p.lock.Unlock()
}

// Not synchronized, lock before calling this method
func (p *peersElector) proceedToPeerElection() {
	if len(p.peers) == 0 {
		p.electedPeer = peer{}
		return
	}
	var max peer
	for _, currentPeer := range p.peers {
		if currentPeer.expiresAt.Before(time.Now()) {
			delete(p.peers, currentPeer.trackerHost)
			continue
		}
		if max.seeders < currentPeer.seeders {
			max = currentPeer
		}
	}

	p.electedPeer = max
}

type SwarmUpdateRequest struct {
	trackerUrl url.URL
	interval   time.Duration
	seeders    int32
	leechers   int32
}

func ErrorSwarmUpdateRequest(trackerUrl url.URL) SwarmUpdateRequest {
	return SwarmUpdateRequest{
		trackerUrl: trackerUrl,
		interval:   1800 * time.Second,
		seeders:    0,
		leechers:   0,
	}
}
func SuccessSwarmUpdateRequest(trackerUrl url.URL, response announcer.AnnounceResponse) SwarmUpdateRequest {
	return SwarmUpdateRequest{
		trackerUrl: trackerUrl,
		interval:   response.Interval * time.Second,
		seeders:    response.Seeders,
		leechers:   response.Leechers,
	}
}

func (s SwarmUpdateRequest) toPeersStats() peer {
	return peer{
		seeders:     s.seeders,
		leechers:    s.leechers,
		trackerHost: s.trackerUrl.Host,
		expiresAt:   time.Now().Add(s.interval * 2),
	}
}
