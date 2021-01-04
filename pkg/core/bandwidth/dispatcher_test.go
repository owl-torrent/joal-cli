package bandwidth

import (
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/nvn1729/congo"
	"github.com/stretchr/testify/assert"
	"strconv"
	"sync"
	"testing"
	"time"
)

type mockedWeightedPool struct {
	getWeights        func() ([]*weightedClaimer, float64)
	removeAllClaimers func()
}

func (m *mockedWeightedPool) GetWeights() (claimers []*weightedClaimer, totalWeight float64) {
	if m.getWeights != nil {
		return m.getWeights()
	}
	return []*weightedClaimer{}, 0
}

func (m *mockedWeightedPool) RemoveAllClaimers() {
	if m.removeAllClaimers != nil {
		m.removeAllClaimers()
	}
}

type mockedRandomSpeedProvider struct {
	bps       int64
	onRefresh func()
}

func (m *mockedRandomSpeedProvider) GetBytesPerSeconds() int64 {
	return m.bps
}

func (m *mockedRandomSpeedProvider) Refresh() {
	if m.onRefresh != nil {
		m.onRefresh()
	}
}

type testSwarm struct {
	seeders  int32
	leechers int32
}

func (s *testSwarm) GetSeeders() int32 {
	return s.seeders
}
func (s *testSwarm) GetLeechers() int32 {
	return s.leechers
}

type dumbSwarm struct {
	seeders  int32
	leechers int32
}

func (s *dumbSwarm) GetSeeders() int32  { return s.seeders }
func (s *dumbSwarm) GetLeechers() int32 { return s.leechers }

type dumbBandwidthClaimable struct {
	infoHash           torrent.InfoHash
	uploaded           int64
	swarm              ISwarm
	onFirstAddUploaded func()
	uploadedWasCalled  bool
	addOnlyOnce        bool
	lock               *sync.Mutex
}

func (bc *dumbBandwidthClaimable) InfoHash() torrent.InfoHash { return bc.infoHash }
func (bc *dumbBandwidthClaimable) AddUploaded(bytes int64) {
	bc.lock.Lock()
	defer bc.lock.Unlock()
	if bc.addOnlyOnce && bc.uploadedWasCalled {
		return
	}
	bc.uploaded += bytes
	if bc.onFirstAddUploaded != nil && !bc.uploadedWasCalled {
		bc.onFirstAddUploaded()
	}
	bc.uploadedWasCalled = true
}
func (bc *dumbBandwidthClaimable) GetSwarm() ISwarm { return bc.swarm }

func TestDispatcher_ShouldBuildFromConfig(t *testing.T) {
	conf := &DispatcherConfig{
		GlobalBandwidthRefreshInterval:           10 * time.Minute,
		IntervalBetweenEachTorrentsSeedIncrement: 1 * time.Minute,
	}
	pool := &mockedWeightedPool{}
	speedProvider := &mockedRandomSpeedProvider{}
	d := NewDispatcher(conf, pool, speedProvider)

	assert.Equal(t, 10*time.Minute, d.(*dispatcher).globalBandwidthRefreshInterval)
	assert.Equal(t, 1*time.Minute, d.(*dispatcher).intervalBetweenEachTorrentsSeedIncrement)
	assert.Equal(t, speedProvider, d.(*dispatcher).randomSpeedProvider)
	assert.Equal(t, pool, d.(*dispatcher).claimerPool)
}

func TestDispatcher_shouldRefreshSpeedProviderOnceOnStart(t *testing.T) {
	latch := congo.NewCountDownLatch(1)

	dispatcher := NewDispatcher(&DispatcherConfig{
		GlobalBandwidthRefreshInterval:           1 * time.Hour,
		IntervalBetweenEachTorrentsSeedIncrement: 1 * time.Hour,
	}, &mockedWeightedPool{}, &mockedRandomSpeedProvider{onRefresh: func() { _ = latch.CountDown() }})

	dispatcher.Start()
	defer dispatcher.Stop()

	if !latch.WaitTimeout(5 * time.Second) {
		t.Fatal("latch has timed out")
	}
}

func TestDispatcher_shouldRefreshSpeedProviderOnTimer(t *testing.T) {
	latch := congo.NewCountDownLatch(4)

	dispatcher := NewDispatcher(&DispatcherConfig{
		GlobalBandwidthRefreshInterval:           1 * time.Millisecond,
		IntervalBetweenEachTorrentsSeedIncrement: 1 * time.Hour,
	}, &mockedWeightedPool{}, &mockedRandomSpeedProvider{onRefresh: func() { _ = latch.CountDown() }})

	dispatcher.Start()
	defer dispatcher.Stop()

	if !latch.WaitTimeout(5 * time.Second) {
		t.Fatal("latch has timed out")
	}
}

func TestDispatcher_shouldDispatchSpeedToRegisteredClaimers(t *testing.T) {
	latch := congo.NewCountDownLatch(1)
	claimer := &dumbBandwidthClaimable{
		infoHash:           metainfo.NewHashFromHex("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"),
		uploaded:           0,
		swarm:              &dumbSwarm{seeders: 100, leechers: 100},
		onFirstAddUploaded: func() { _ = latch.CountDown() },
		addOnlyOnce:        true,
		lock:               &sync.Mutex{},
	}

	pool := &mockedWeightedPool{
		getWeights: func() ([]*weightedClaimer, float64) {
			return []*weightedClaimer{{
				IBandwidthClaimable: claimer,
				weight:              152,
			}}, 152
		},
	}
	dispatcher := NewDispatcher(&DispatcherConfig{
		GlobalBandwidthRefreshInterval:           1 * time.Hour,
		IntervalBetweenEachTorrentsSeedIncrement: 1 * time.Millisecond,
	}, pool, &mockedRandomSpeedProvider{bps: 10000000})

	dispatcher.Start()
	defer dispatcher.Stop()
	if !latch.WaitTimeout(5 * time.Second) {
		t.Fatal("latch has timed out")
	}

	assert.Greater(t, claimer.uploaded, int64(0))
}

func TestDispatcher_shouldDispatchBasedOnWeight(t *testing.T) {
	wg := sync.WaitGroup{}
	claimer1 := &dumbBandwidthClaimable{
		infoHash:           metainfo.NewHashFromHex("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"),
		uploaded:           0,
		swarm:              &dumbSwarm{seeders: 100, leechers: 2500},
		onFirstAddUploaded: func() { wg.Done() },
		addOnlyOnce:        true,
		lock:               &sync.Mutex{},
	}
	claimer2 := &dumbBandwidthClaimable{
		infoHash: metainfo.NewHashFromHex("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"),
		uploaded: 0,
		swarm:    &dumbSwarm{seeders: 10, leechers: 5},
		onFirstAddUploaded: func() {
			wg.Done()
		},
		addOnlyOnce: true,
		lock:        &sync.Mutex{},
	}

	pool := &mockedWeightedPool{
		getWeights: func() ([]*weightedClaimer, float64) {
			return []*weightedClaimer{{
				IBandwidthClaimable: claimer1,
				weight:              152,
			}, {
				IBandwidthClaimable: claimer2,
				weight:              15.2,
			}}, 167.2
		},
	}
	dispatcher := NewDispatcher(&DispatcherConfig{
		GlobalBandwidthRefreshInterval:           1 * time.Hour,
		IntervalBetweenEachTorrentsSeedIncrement: 1 * time.Millisecond,
	}, pool, &mockedRandomSpeedProvider{bps: 10000000})

	wg.Add(2)

	dispatcher.Start()
	defer dispatcher.Stop()
	wg.Wait()
	assert.Greater(t, claimer1.uploaded, claimer2.uploaded)
}

func TestDispatcher_shouldDispatchBasedOnWeightFiftyFifty(t *testing.T) {
	wg := sync.WaitGroup{}
	claimer1 := &dumbBandwidthClaimable{
		infoHash:           metainfo.NewHashFromHex("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"),
		uploaded:           0,
		swarm:              &dumbSwarm{seeders: 100, leechers: 2500},
		onFirstAddUploaded: func() { wg.Done() },
		addOnlyOnce:        true,
		lock:               &sync.Mutex{},
	}
	claimer2 := &dumbBandwidthClaimable{
		infoHash:           metainfo.NewHashFromHex("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"),
		uploaded:           0,
		swarm:              &dumbSwarm{seeders: 100, leechers: 2500},
		onFirstAddUploaded: func() { wg.Done() },
		addOnlyOnce:        true,
		lock:               &sync.Mutex{},
	}

	wg.Add(2)
	pool := &mockedWeightedPool{
		getWeights: func() ([]*weightedClaimer, float64) {
			return []*weightedClaimer{{
				IBandwidthClaimable: claimer1,
				weight:              152,
			}, {
				IBandwidthClaimable: claimer2,
				weight:              152,
			}}, 304
		},
	}
	dispatcher := NewDispatcher(&DispatcherConfig{
		GlobalBandwidthRefreshInterval:           1 * time.Hour,
		IntervalBetweenEachTorrentsSeedIncrement: 1 * time.Millisecond,
	}, pool, &mockedRandomSpeedProvider{bps: 10000000})

	dispatcher.Start()
	wg.Wait()
	dispatcher.Stop()
	assert.Equal(t, claimer1.uploaded, claimer2.uploaded)
}

func TestDispatcher_shouldNotDispatchIfNoWeight(t *testing.T) {
	wg := sync.WaitGroup{}
	claimer1 := &dumbBandwidthClaimable{
		infoHash:           metainfo.NewHashFromHex("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"),
		uploaded:           0,
		swarm:              &dumbSwarm{seeders: 0, leechers: 0},
		onFirstAddUploaded: func() { wg.Done() },
		addOnlyOnce:        true,
		lock:               &sync.Mutex{},
	}
	claimer2 := &dumbBandwidthClaimable{
		infoHash:           metainfo.NewHashFromHex("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"),
		uploaded:           0,
		swarm:              &dumbSwarm{seeders: 100, leechers: 2500},
		onFirstAddUploaded: func() { wg.Done() },
		addOnlyOnce:        true,
		lock:               &sync.Mutex{},
	}

	pool := &mockedWeightedPool{
		getWeights: func() ([]*weightedClaimer, float64) {
			return []*weightedClaimer{{
				IBandwidthClaimable: claimer1,
				weight:              0,
			}, {
				IBandwidthClaimable: claimer2,
				weight:              152,
			}}, 152
		},
	}
	dispatcher := NewDispatcher(&DispatcherConfig{
		GlobalBandwidthRefreshInterval:           1 * time.Hour,
		IntervalBetweenEachTorrentsSeedIncrement: 1 * time.Millisecond,
	}, pool, &mockedRandomSpeedProvider{bps: 10000000})

	wg.Add(2)

	dispatcher.Start()
	wg.Wait()
	dispatcher.Stop()
	assert.Zero(t, claimer1.uploaded)
	assert.Greater(t, claimer2.uploaded, int64(0))
}

func TestDispatcher_ShouldResetWeightPoolOnStop(t *testing.T) {
	hasRemovedAll := false
	pool := &mockedWeightedPool{
		removeAllClaimers: func() {
			hasRemovedAll = true
		},
	}
	dispatcher := NewDispatcher(&DispatcherConfig{
		GlobalBandwidthRefreshInterval:           1 * time.Hour,
		IntervalBetweenEachTorrentsSeedIncrement: 1 * time.Millisecond,
	}, pool, &mockedRandomSpeedProvider{bps: 10000000})

	dispatcher.Start()
	dispatcher.Stop()

	assert.True(t, hasRemovedAll)
}

type mockedBandwidthClaimable struct {
	infoHash    torrent.InfoHash
	uploaded    int64
	swarm       ISwarm
	addUploaded func(bytes int64)
}

func (c *mockedBandwidthClaimable) InfoHash() torrent.InfoHash {
	return c.infoHash
}

func (c *mockedBandwidthClaimable) AddUploaded(bytes int64) {
	if c.addUploaded != nil {
		c.addUploaded(bytes)
		return
	}
	c.uploaded += bytes
}

func (c *mockedBandwidthClaimable) GetSwarm() ISwarm {
	return c.swarm
}

func TestDispatcher_ShouldWorkWithTremendousAmountOfClaimers(t *testing.T) {
	numberOfClaimers := 5000

	weightedClaimerSlice := make([]*weightedClaimer, numberOfClaimers)
	totalWeight := 0.0
	pool := &mockedWeightedPool{
		getWeights: func() ([]*weightedClaimer, float64) {
			return weightedClaimerSlice, totalWeight
		},
	}

	latch := congo.NewCountDownLatch(uint(numberOfClaimers) * 20)
	for i := 0; i < numberOfClaimers; i++ {
		claimer := &mockedBandwidthClaimable{
			infoHash: metainfo.NewHashFromHex(fmt.Sprintf("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA%d", i)[len(strconv.Itoa(i)):]),
			uploaded: 0,
			swarm: &dumbSwarm{
				seeders:  10,
				leechers: 10,
			},
		}
		claimer.addUploaded = func(bytes int64) {
			claimer.uploaded += bytes
			_ = latch.CountDown()
		}
		weightedClaimerSlice[i] = &weightedClaimer{
			IBandwidthClaimable: claimer,
			weight:              100,
		}
		totalWeight += 100
	}

	d := NewDispatcher(&DispatcherConfig{
		GlobalBandwidthRefreshInterval:           1 * time.Hour,
		IntervalBetweenEachTorrentsSeedIncrement: 1 * time.Millisecond,
	}, pool, &mockedRandomSpeedProvider{bps: 100})

	d.Start()
	defer d.Stop()

	if !latch.WaitTimeout(5 * time.Second) {
		t.Fatal("timeout")
	}
}
