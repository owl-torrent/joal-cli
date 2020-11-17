package bandwidth

import (
	"github.com/anacrolix/torrent"
	"sync"
	"time"
)

type IBandwidthClaimable interface {
	InfoHash() torrent.InfoHash
	AddUploaded(bytes int64)
	// May return nil
	GetSwarm() ISwarm
}
type ISwarm interface {
	GetSeeders() int32
	GetLeechers() int32
}

type IDispatcher interface {
	Start()
	Stop()
	ClaimOrUpdate(claimer IBandwidthClaimable)
	Release(claimer IBandwidthClaimable)
}

func newDispatcher(conf *DispatcherConfig, rsp iRandomSpeedProvider) IDispatcher {
	return &dispatcher{
		globalBandwidthRefreshInterval:           conf.GlobalBandwidthRefreshInterval,
		intervalBetweenEachTorrentsSeedIncrement: conf.IntervalBetweenEachTorrentsSeedIncrement,
		randomSpeedProvider:                      rsp,
		claimers:                                 make(map[torrent.InfoHash]weightedClaimer),
		totalWeight:                              0,
		isRunning:                                false,
		stopping:                                 make(chan chan struct{}),
		lock:                                     &sync.RWMutex{},
	}
}

type claimerWeight = float64
type weightedClaimer struct {
	IBandwidthClaimable
	weight claimerWeight
}

type dispatcher struct {
	globalBandwidthRefreshInterval           time.Duration
	intervalBetweenEachTorrentsSeedIncrement time.Duration
	randomSpeedProvider                      iRandomSpeedProvider
	claimers                                 map[torrent.InfoHash]weightedClaimer
	totalWeight                              float64

	isRunning bool
	stopping  chan chan struct{}
	lock      *sync.RWMutex
}

func (d *dispatcher) Start() {
	d.lock.Lock()
	defer d.lock.Unlock()
	if d.isRunning {
		return
	}
	d.isRunning = true

	go func() {
		d.randomSpeedProvider.Refresh()

		globalBandwidthRefreshTicker := time.NewTicker(d.globalBandwidthRefreshInterval)
		timeToAddSeedToClaimers := time.NewTicker(d.intervalBetweenEachTorrentsSeedIncrement)
		secondsBetweenLoops := d.intervalBetweenEachTorrentsSeedIncrement.Seconds()

		for {
			select {
			case <-globalBandwidthRefreshTicker.C:
				d.randomSpeedProvider.Refresh()
			case <-timeToAddSeedToClaimers.C:
				if d.totalWeight == 0 {
					continue
				}
				bytesToDispatch := float64(d.randomSpeedProvider.GetBytesPerSeconds()) * secondsBetweenLoops
				d.lock.RLock()
				for _, claimer := range d.claimers {
					percentOfSpeedToAssign := claimer.weight / d.totalWeight
					claimer.AddUploaded(int64(bytesToDispatch * percentOfSpeedToAssign))
				}
				d.lock.RUnlock()
			case doneStopping := <-d.stopping:
				timeToAddSeedToClaimers.Stop()
				globalBandwidthRefreshTicker.Stop()
				doneStopping <- struct{}{}
				return
			}
		}
	}()
}

func (d *dispatcher) Stop() {
	d.lock.Lock()
	defer d.lock.Unlock()
	if !d.isRunning {
		return
	}
	d.isRunning = false

	doneStopping := make(chan struct{})
	d.stopping <- doneStopping

	<-doneStopping
}

// Register a IBandwidthClaimable as a bandwidth client. Will update his uploaded stats on a timer and the amount of uploaded given depend on this ISwarm of the IBandwidthClaimable.
// If called with an already known IBandwidthClaimable, re-calculate his bandwidth attribution based on his ISwarm. Basically this methods should be called every time the IBandwidthClaimable receives new Peers from the tracker.
func (d *dispatcher) ClaimOrUpdate(claimer IBandwidthClaimable) {
	d.lock.RLock()
	previousClaimer, previousClaimerExists := d.claimers[claimer.InfoHash()]
	if previousClaimerExists {
		d.totalWeight -= previousClaimer.weight
	}
	d.lock.RUnlock()

	d.lock.Lock()
	defer d.lock.Unlock()
	weight := calculateWeight(claimer.GetSwarm())
	d.claimers[claimer.InfoHash()] = weightedClaimer{
		IBandwidthClaimable: claimer,
		weight:              weight,
	}
	d.totalWeight += weight
}

// Unregister a IBandwidthClaimable. After being released a IBandwidthClaimable wont receive any more bandwidth
func (d *dispatcher) Release(claimer IBandwidthClaimable) {
	d.lock.RLock()
	previousClaimerWeight, exists := d.claimers[claimer.InfoHash()]
	d.lock.RUnlock()

	if exists {
		d.lock.Lock()
		defer d.lock.Unlock()

		d.totalWeight -= previousClaimerWeight.weight
		delete(d.claimers, claimer.InfoHash())
	}
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
