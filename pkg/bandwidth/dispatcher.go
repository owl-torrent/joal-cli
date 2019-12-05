package bandwidth

import (
	"github.com/anacrolix/torrent"
	"github.com/anthonyraymond/joal-cli/pkg/utils"
	"sync"
	"time"
)

type ISwarm interface {
	GetSeeders() int32
	GetLeechers() int32
}
type IBandwidthClaimable interface {
	InfoHash() *torrent.InfoHash
	AddUploaded(bytes int64)
	// May return nil
	GetSwarm() ISwarm
}

type IDispatcher interface {
	Start()
	Stop()
	ClaimOrUpdate(claimer IBandwidthClaimable)
	Release(claimer IBandwidthClaimable)
}

type Weight = float64

type dispatcher struct {
	speedProviderUpdateInterval time.Duration
	dispatcherUpdateInterval    time.Duration
	quit                        chan int
	randomSpeedProvider         IRandomSpeedProvider
	claimers                    map[IBandwidthClaimable]Weight
	totalWeight                 float64
	lock                        *sync.RWMutex
}

func DispatcherNew(randomSpeedProvider IRandomSpeedProvider) IDispatcher {
	return &dispatcher{
		speedProviderUpdateInterval: 20 * time.Minute,
		dispatcherUpdateInterval:    5 * time.Second,
		randomSpeedProvider:         randomSpeedProvider,
		claimers:                    make(map[IBandwidthClaimable]Weight),
		totalWeight:                 0,
		lock:                        &sync.RWMutex{},
	}
}

func (d *dispatcher) Start() {
	d.quit = make(chan int)
	go func() {
		speedProviderChan := utils.Every(d.speedProviderUpdateInterval, func() { d.randomSpeedProvider.Refresh() })
		defer close(speedProviderChan)
		ticker := time.NewTicker(d.dispatcherUpdateInterval)
		for {
			select {
			case <-ticker.C:
				if d.totalWeight == 0 {
					continue
				}
				d.lock.RLock()
				for claimer, weight := range d.claimers {
					bytesToDispatch := float64(d.randomSpeedProvider.GetBytesPerSeconds()) * d.dispatcherUpdateInterval.Seconds()
					percentOfSpeedToAssign := weight / d.totalWeight
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

	previousClaimerWeight, exists := d.claimers[claimer]
	if exists {
		d.totalWeight -= previousClaimerWeight
	}

	d.claimers[claimer] = calculateWeight(claimer.GetSwarm())
	d.totalWeight += d.claimers[claimer]
}

// Unregister a IBandwidthClaimable. After being released a IBandwidthClaimable wont receive any more bandwidth
func (d *dispatcher) Release(claimer IBandwidthClaimable) {
	d.lock.Lock()
	defer d.lock.Unlock()

	previousClaimerWeight, exists := d.claimers[claimer]
	if exists {
		d.totalWeight -= previousClaimerWeight
		delete(d.claimers, claimer)
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
