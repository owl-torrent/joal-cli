package bandwidth

import (
	"github.com/anacrolix/torrent"
	"github.com/anthonyraymond/joal-cli/internal/core/broadcast"
	"sync"
)

type claimerWeight = float64
type weightedClaimer struct {
	IBandwidthClaimable
	weight claimerWeight
}

type IBandwidthWeightedClaimerPool interface {
	// Returns a list of all the weightedClaimers alongs with the sum of all their weights. The slice must be access safe and is most likely a copy of the underlying struct storage
	GetWeights() (claimers []*weightedClaimer, totalWeight float64)
	RemoveAllClaimers()
}

type IBandwidthClaimerPool interface {
	AddOrUpdate(claimer IBandwidthClaimable)
	RemoveFromPool(claimer IBandwidthClaimable)
}

type claimerPool struct {
	claimers    map[torrent.InfoHash]weightedClaimer
	totalWeight float64
	lock        *sync.RWMutex
}

func NewWeightedClaimerPool() *claimerPool {
	return &claimerPool{
		claimers:    make(map[torrent.InfoHash]weightedClaimer),
		totalWeight: 0,
		lock:        &sync.RWMutex{},
	}
}

func (c *claimerPool) GetWeights() (claimers []*weightedClaimer, totalWeight float64) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	// return a 'copy' so that user of claimers wont run into thread race issues
	res := make([]*weightedClaimer, len(c.claimers))
	i := 0
	for _, claimer := range c.claimers {
		res[i] = &claimer
		i++
	}
	return res, c.totalWeight
}

func (c *claimerPool) RemoveAllClaimers() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.claimers = make(map[torrent.InfoHash]weightedClaimer)
	c.totalWeight = 0
}

func (c *claimerPool) AddOrUpdate(claimer IBandwidthClaimable) {
	c.lock.Lock()
	defer c.lock.Unlock()

	previousClaimer, previousClaimerExists := c.claimers[claimer.InfoHash()]
	if previousClaimerExists {
		c.totalWeight -= previousClaimer.weight
	}

	weight := calculateWeight(claimer.GetSwarm())
	c.claimers[claimer.InfoHash()] = weightedClaimer{
		IBandwidthClaimable: claimer,
		weight:              weight,
	}
	c.totalWeight += weight

	broadcastWeights(c)
}

func (c *claimerPool) RemoveFromPool(claimer IBandwidthClaimable) {
	c.lock.Lock()
	defer c.lock.Unlock()

	previousClaimerWeight, exists := c.claimers[claimer.InfoHash()]

	if !exists {
		return
	}
	c.totalWeight -= previousClaimerWeight.weight
	delete(c.claimers, claimer.InfoHash())
	broadcastWeights(c)
}

func calculateWeight(swarm ISwarm) float64 {
	if swarm == nil || swarm.GetSeeders() == 0 || swarm.GetLeechers() == 0 {
		return 0
	}
	leechersRatio := float64(swarm.GetLeechers()) / float64(swarm.GetSeeders()+swarm.GetLeechers())
	if leechersRatio == 0.0 {
		return 0
	}

	return leechersRatio * 100.0 * (float64(swarm.GetSeeders()) * leechersRatio) * (float64(swarm.GetLeechers()) / float64(swarm.GetSeeders()))
}

func broadcastWeights(pool *claimerPool) {
	weightMap := make(map[torrent.InfoHash]float64, len(pool.claimers))
	for infohash, weightedClaimer := range pool.claimers {
		weightMap[infohash] = weightedClaimer.weight
	}
	broadcast.EmitBandwidthWeightHasChanged(broadcast.BandwidthWeightHasChangedEvent{
		TotalWeight:    pool.totalWeight,
		TorrentWeights: weightMap,
	})
}
