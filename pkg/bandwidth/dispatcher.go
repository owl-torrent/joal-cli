package bandwidth

import (
	"github.com/anacrolix/torrent"
	"github.com/anthonyraymond/joal-cli/pkg/utils/timeutils"
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
		claimers:                                 make(map[torrent.InfoHash]weigthedClaimer),
		totalWeight:                              0,
		lock:                                     &sync.RWMutex{},
	}
}

type claimerWeight = float64
type weigthedClaimer struct {
	IBandwidthClaimable
	weight claimerWeight
}

type dispatcher struct {
	globalBandwidthRefreshInterval           time.Duration
	intervalBetweenEachTorrentsSeedIncrement time.Duration
	quit                                     chan int
	randomSpeedProvider                      iRandomSpeedProvider
	claimers                                 map[torrent.InfoHash]weigthedClaimer
	totalWeight                              float64
	lock                                     *sync.RWMutex
}

func (d *dispatcher) Start() {
	// TODO: rewrite properly with channels instead of timeutils.every
	d.quit = make(chan int)
	go func() {
		d.randomSpeedProvider.Refresh()
		speedProviderChan := timeutils.Every(d.globalBandwidthRefreshInterval, func() { d.randomSpeedProvider.Refresh() })
		defer close(speedProviderChan)
		ticker := time.NewTicker(d.intervalBetweenEachTorrentsSeedIncrement)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if d.totalWeight == 0 {
					continue
				}
				d.lock.RLock()
				bytesToDispatch := float64(d.randomSpeedProvider.GetBytesPerSeconds()) * d.intervalBetweenEachTorrentsSeedIncrement.Seconds()
				for _, claimer := range d.claimers {
					percentOfSpeedToAssign := claimer.weight / d.totalWeight
					claimer.AddUploaded(int64(bytesToDispatch * percentOfSpeedToAssign))
				}
				d.lock.RUnlock()
			case <-d.quit:
				return
			}
		}
	}()
}
func (d *dispatcher) Stop() {
	close(d.quit)
}

// Register a IBandwidthClaimable as a bandwidth client. Will update his uploaded stats on a timer and the amount of uploaded given depend on this ISwarm of the IBandwidthClaimable.
// If called with an already known IBandwidthClaimable, re-calculate his bandwidth attribution based on his ISwarm. Basically this methods should be called every time the IBandwidthClaimable receives new Peers from the tracker.
func (d *dispatcher) ClaimOrUpdate(claimer IBandwidthClaimable) {
	d.lock.Lock()
	defer d.lock.Unlock()

	previousClaimer, previousClaimerExists := d.claimers[claimer.InfoHash()]
	if previousClaimerExists {
		d.totalWeight -= previousClaimer.weight
	}

	weight := calculateWeight(claimer.GetSwarm())
	d.claimers[claimer.InfoHash()] = weigthedClaimer{
		IBandwidthClaimable: claimer,
		weight:              weight,
	}
	d.totalWeight += weight
}

// Unregister a IBandwidthClaimable. After being released a IBandwidthClaimable wont receive any more bandwidth
func (d *dispatcher) Release(claimer IBandwidthClaimable) {
	d.lock.Lock()
	defer d.lock.Unlock()

	previousClaimerWeight, exists := d.claimers[claimer.InfoHash()]
	if exists {
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
