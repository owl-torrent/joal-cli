package bandwidth

import (
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anthonyraymond/joal-cli/internal/core"
	"github.com/anthonyraymond/joal-cli/internal/core/broadcast"
	"github.com/anthonyraymond/joal-cli/internal/core/logs"
	"github.com/anthonyraymond/joal-cli/internal/utils/dataunit"
	"go.uber.org/zap"
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
}

func NewDispatcher(conf *core.DispatcherConfig, pool IBandwidthWeightedClaimerPool, rsp iRandomSpeedProvider) IDispatcher {
	return &dispatcher{
		globalBandwidthRefreshInterval:           conf.GlobalBandwidthRefreshInterval,
		intervalBetweenEachTorrentsSeedIncrement: conf.IntervalBetweenEachTorrentsSeedIncrement,
		claimerPool:                              pool,
		randomSpeedProvider:                      rsp,
		isRunning:                                false,
		stopping:                                 make(chan chan struct{}),
		lock:                                     &sync.RWMutex{},
	}
}

type dispatcher struct {
	globalBandwidthRefreshInterval           time.Duration
	intervalBetweenEachTorrentsSeedIncrement time.Duration
	randomSpeedProvider                      iRandomSpeedProvider
	claimerPool                              IBandwidthWeightedClaimerPool
	isRunning                                bool
	stopping                                 chan chan struct{}
	lock                                     *sync.RWMutex
}

func (d *dispatcher) Start() {
	d.lock.Lock()
	defer d.lock.Unlock()
	if d.isRunning {
		return
	}
	d.isRunning = true

	log := logs.GetLogger()
	go func() {
		d.randomSpeedProvider.Refresh()
		log.Info("bandwidth dispatcher: started",
			zap.String("available-bandwidth", fmt.Sprintf("%s/s", dataunit.ByteCountSI(d.randomSpeedProvider.GetBytesPerSeconds()))),
		)

		globalBandwidthRefreshTicker := time.NewTicker(d.globalBandwidthRefreshInterval)
		broadcast.EmitGlobalBandwidthChanged(broadcast.GlobalBandwidthChangedEvent{AvailableBandwidth: d.randomSpeedProvider.GetBytesPerSeconds()})
		timeToAddSeedToClaimers := time.NewTicker(d.intervalBetweenEachTorrentsSeedIncrement)
		secondsBetweenLoops := d.intervalBetweenEachTorrentsSeedIncrement.Seconds()

		for {
			select {
			case <-globalBandwidthRefreshTicker.C:
				d.randomSpeedProvider.Refresh()
				broadcast.EmitGlobalBandwidthChanged(broadcast.GlobalBandwidthChangedEvent{AvailableBandwidth: d.randomSpeedProvider.GetBytesPerSeconds()})
				log.Info("bandwidth dispatcher: refreshed available bandwidth",
					zap.String("available-bandwidth", fmt.Sprintf("%s/s", dataunit.ByteCountSI(d.randomSpeedProvider.GetBytesPerSeconds()))),
				)
			case <-timeToAddSeedToClaimers.C:
				claimers, totalWeight := d.claimerPool.GetWeights()

				if totalWeight == 0 {
					continue
				}

				bytesToDispatch := float64(d.randomSpeedProvider.GetBytesPerSeconds()) * secondsBetweenLoops
				for _, claimer := range claimers {
					percentOfSpeedToAssign := claimer.weight / totalWeight
					claimer.AddUploaded(int64(bytesToDispatch * percentOfSpeedToAssign))
				}
			case doneStopping := <-d.stopping:
				timeToAddSeedToClaimers.Stop()
				globalBandwidthRefreshTicker.Stop()
				d.claimerPool.RemoveAllClaimers()
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

	log := logs.GetLogger()
	log.Info("bandwidth dispatcher: stopping")

	doneStopping := make(chan struct{})
	d.stopping <- doneStopping

	<-doneStopping
	log.Info("bandwidth dispatcher: stopped")
}
