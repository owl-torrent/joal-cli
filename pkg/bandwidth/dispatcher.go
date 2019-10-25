package bandwidth

import (
	"context"
	"github.com/anacrolix/torrent"
	"sync"
)

type ISwarm interface {
	getSeeders() uint64
	getLeechers() uint64
}
type IBandwidthClaimable interface {
	InfoHash() *torrent.InfoHash
	AddUploaded(bytes uint64)
	// May return nil
	getSwarm() ISwarm
}

type IDispatcher interface {
	Start()
	Stop(ctx context.Context)
	Claim(claimer IBandwidthClaimable)
	Release(claimer IBandwidthClaimable)
}

type Weight = float64

type Dispatcher struct {
	claimers    map[IBandwidthClaimable]Weight
	totalWeight float64
	lock        *sync.RWMutex
}

func (d *Dispatcher) Claim(claimer IBandwidthClaimable) {
	d.lock.Lock()
	defer d.lock.Unlock()

	previousClaimerWeight, exists := d.claimers[claimer]
	if exists {
		d.totalWeight -= previousClaimerWeight
	}

	d.claimers[claimer] = calculateWeight(claimer.getSwarm())
	d.totalWeight += d.claimers[claimer]
}

func (d *Dispatcher) Release(claimer IBandwidthClaimable) {
	d.lock.Lock()
	defer d.lock.Unlock()

	previousClaimerWeight, exists := d.claimers[claimer]
	if exists {
		d.totalWeight -= previousClaimerWeight
		delete(d.claimers, claimer)
	}
}

func calculateWeight(swarm ISwarm) float64 {
	if swarm == nil || swarm.getSeeders() == 0 || swarm.getLeechers() == 0 {
		return 0
	}
	leechersRatio := float64(swarm.getLeechers()) / float64(swarm.getSeeders() + swarm.getLeechers())
	if leechersRatio == 0.0 {
		return 0
	}

	return leechersRatio * 100.0 * (float64(swarm.getSeeders()) * leechersRatio) * (float64(swarm.getLeechers()) / float64(swarm.getSeeders()));
}
