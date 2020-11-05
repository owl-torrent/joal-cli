package bandwidth

import (
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/nvn1729/congo"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

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

func Test_calculateWeightShouldNeverGoBelowZero(t *testing.T) {
	type args struct {
		swarm ISwarm
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{name: "with0Leechers0Seeders", args: args{swarm: &testSwarm{seeders: 0, leechers: 0}}, want: 0},
		{name: "with0Leechers1Seeders", args: args{swarm: &testSwarm{seeders: 1, leechers: 0}}, want: 0},
		{name: "with1Leechers0Seeders", args: args{swarm: &testSwarm{seeders: 0, leechers: 1}}, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateWeight(tt.args.swarm); got != tt.want {
				t.Errorf("calculateWeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calculateWeightShouldNotFailIfSwarmIsNil(t *testing.T) {
	assert.Zero(t, calculateWeight(nil))
}

func Test_calculateWeightShouldPromoteSwarmWithMoreLeechers(t *testing.T) {
	first := calculateWeight(&testSwarm{seeders: 10, leechers: 10})
	second := calculateWeight(&testSwarm{seeders: 10, leechers: 30})
	third := calculateWeight(&testSwarm{seeders: 10, leechers: 100})
	fourth := calculateWeight(&testSwarm{seeders: 10, leechers: 200})
	assert.Greater(t, fourth, third, "should be greater")
	assert.Greater(t, third, second, "should be greater")
	assert.Greater(t, second, first, "should be greater")
}

func Test_calculateWeightShouldProvidePreciseValues(t *testing.T) {
	assert.Zero(t, calculateWeight(&testSwarm{seeders: 0, leechers: 1}))
	assert.Zero(t, calculateWeight(&testSwarm{seeders: 1, leechers: 0}))
	assert.Zero(t, calculateWeight(&testSwarm{seeders: 0, leechers: 100}))
	assert.Zero(t, calculateWeight(&testSwarm{seeders: 0, leechers: 0}))
	assert.Equal(t, 25.0, calculateWeight(&testSwarm{seeders: 1, leechers: 1}))
	assert.Equal(t, 50000.0, calculateWeight(&testSwarm{seeders: 2000, leechers: 2000}))
	assert.InDelta(t, float64(11.1), calculateWeight(&testSwarm{seeders: 2, leechers: 1}), 0.1)
	assert.InDelta(t, float64(0.104058273), calculateWeight(&testSwarm{seeders: 30, leechers: 1}), 0.00000001)
	assert.InDelta(t, float64(9611.687812), calculateWeight(&testSwarm{seeders: 2, leechers: 100}), 0.0001)
	assert.InDelta(t, float64(73.01243916), calculateWeight(&testSwarm{seeders: 2000, leechers: 150}), 0.00001)
	assert.InDelta(t, float64(173066.5224), calculateWeight(&testSwarm{seeders: 150, leechers: 2000}), 0.01)
	assert.InDelta(t, float64(184911.2426), calculateWeight(&testSwarm{seeders: 80, leechers: 2000}), 0.1)
}

type DumbSwarm struct {
	seeders  int32
	leechers int32
}

func (s *DumbSwarm) GetSeeders() int32  { return s.seeders }
func (s *DumbSwarm) GetLeechers() int32 { return s.leechers }

type DumbBandwidthClaimable struct {
	infoHash           torrent.InfoHash
	uploaded           int64
	swarm              ISwarm
	onFirstAddUploaded func()
	uploadedWasCalled  bool
	addOnlyOnce        bool
	lock               *sync.Mutex
}

func (bc *DumbBandwidthClaimable) InfoHash() torrent.InfoHash { return bc.infoHash }
func (bc *DumbBandwidthClaimable) AddUploaded(bytes int64) {
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
func (bc *DumbBandwidthClaimable) GetSwarm() ISwarm { return bc.swarm }

func TestDispatcher_shouldRefreshSpeedProviderOnceOnStart(t *testing.T) {
	latch := congo.NewCountDownLatch(1)

	dispatcher := &dispatcher{
		speedProviderUpdateInterval: 1 * time.Hour,
		dispatcherUpdateInterval:    1 * time.Millisecond,
		randomSpeedProvider:         &mockedRandomSpeedProvider{onRefresh: func() { _ = latch.CountDown() }},
		claimers:                    make(map[IBandwidthClaimable]Weight),
		totalWeight:                 0,
		lock:                        &sync.RWMutex{},
	}

	dispatcher.Start()
	defer dispatcher.Stop()

	if !latch.WaitTimeout(5 * time.Second) {
		t.Fatal("latch has timed out")
	}
}

func TestDispatcher_shouldRefreshSpeedProviderOnTimer(t *testing.T) {
	latch := congo.NewCountDownLatch(4)

	dispatcher := &dispatcher{
		speedProviderUpdateInterval: 1 * time.Millisecond,
		dispatcherUpdateInterval:    1 * time.Millisecond,
		randomSpeedProvider:         &mockedRandomSpeedProvider{onRefresh: func() { _ = latch.CountDown() }},
		claimers:                    make(map[IBandwidthClaimable]Weight),
		totalWeight:                 0,
		lock:                        &sync.RWMutex{},
	}

	dispatcher.Start()
	defer dispatcher.Stop()

	if !latch.WaitTimeout(5 * time.Second) {
		t.Fatal("latch has timed out")
	}
}

type DumbStaticSpeedProvider struct {
	speed        int64
	refreshCount int
}

func (s *DumbStaticSpeedProvider) GetBytesPerSeconds() int64 { return s.speed }
func (s *DumbStaticSpeedProvider) Refresh()                  { s.refreshCount += 1 }

func TestDispatcher_shouldDispatchSpeedToRegisteredClaimers(t *testing.T) {
	dispatcher := &dispatcher{
		speedProviderUpdateInterval: 1 * time.Hour,
		dispatcherUpdateInterval:    1 * time.Millisecond,
		randomSpeedProvider:         &mockedRandomSpeedProvider{bps: 10000000},
		claimers:                    make(map[IBandwidthClaimable]Weight),
		totalWeight:                 0,
		lock:                        &sync.RWMutex{},
	}

	latch := congo.NewCountDownLatch(1)
	ih1 := metainfo.NewHashFromHex("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	claimer := &DumbBandwidthClaimable{
		infoHash:           ih1,
		uploaded:           0,
		swarm:              &DumbSwarm{seeders: 100, leechers: 100},
		onFirstAddUploaded: func() { latch.CountDown() },
		addOnlyOnce:        true,
		lock:               &sync.Mutex{},
	}

	dispatcher.ClaimOrUpdate(claimer)

	dispatcher.Start()
	defer dispatcher.Stop()
	if !latch.WaitTimeout(5 * time.Second) {
		t.Fatal("latch has timed out")
	}

	assert.Greater(t, claimer.uploaded, int64(0))
}

func TestDispatcher_shouldDispatchBasedOnWeight(t *testing.T) {
	dispatcher := &dispatcher{
		speedProviderUpdateInterval: 1 * time.Hour,
		dispatcherUpdateInterval:    1 * time.Millisecond,
		randomSpeedProvider:         &DumbStaticSpeedProvider{speed: 10000000},
		claimers:                    make(map[IBandwidthClaimable]Weight),
		totalWeight:                 0,
		lock:                        &sync.RWMutex{},
	}
	wg := sync.WaitGroup{}
	ih1 := metainfo.NewHashFromHex("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	claimer1 := &DumbBandwidthClaimable{
		infoHash:           ih1,
		uploaded:           0,
		swarm:              &DumbSwarm{seeders: 100, leechers: 2500},
		onFirstAddUploaded: func() { wg.Done() },
		addOnlyOnce:        true,
		lock:               &sync.Mutex{},
	}
	ih2 := metainfo.NewHashFromHex("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB")
	claimer2 := &DumbBandwidthClaimable{
		infoHash:           ih2,
		uploaded:           0,
		swarm:              &DumbSwarm{seeders: 10, leechers: 5},
		onFirstAddUploaded: func() { wg.Done() },
		addOnlyOnce:        true,
		lock:               &sync.Mutex{},
	}

	wg.Add(2)
	dispatcher.ClaimOrUpdate(claimer1)
	dispatcher.ClaimOrUpdate(claimer2)

	dispatcher.Start()
	wg.Wait()
	dispatcher.Stop()
	assert.Greater(t, claimer1.uploaded, claimer2.uploaded)
}

func TestDispatcher_shouldDispatchBasedOnWeightFiftyFifty(t *testing.T) {
	dispatcher := &dispatcher{
		speedProviderUpdateInterval: 1 * time.Hour,
		dispatcherUpdateInterval:    1 * time.Millisecond,
		randomSpeedProvider:         &DumbStaticSpeedProvider{speed: 10000000},
		claimers:                    make(map[IBandwidthClaimable]Weight),
		totalWeight:                 0,
		lock:                        &sync.RWMutex{},
	}
	wg := sync.WaitGroup{}
	ih1 := metainfo.NewHashFromHex("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	claimer1 := &DumbBandwidthClaimable{
		infoHash:           ih1,
		uploaded:           0,
		swarm:              &DumbSwarm{seeders: 100, leechers: 2500},
		onFirstAddUploaded: func() { wg.Done() },
		addOnlyOnce:        true,
		lock:               &sync.Mutex{},
	}
	ih2 := metainfo.NewHashFromHex("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB")
	claimer2 := &DumbBandwidthClaimable{
		infoHash:           ih2,
		uploaded:           0,
		swarm:              &DumbSwarm{seeders: 100, leechers: 2500},
		onFirstAddUploaded: func() { wg.Done() },
		addOnlyOnce:        true,
		lock:               &sync.Mutex{},
	}

	wg.Add(2)
	dispatcher.ClaimOrUpdate(claimer1)
	dispatcher.ClaimOrUpdate(claimer2)

	dispatcher.Start()
	wg.Wait()
	dispatcher.Stop()
	assert.Equal(t, claimer1.uploaded, claimer2.uploaded)
}

func TestDispatcher_shouldNotDispatchIfNoPeers(t *testing.T) {
	dispatcher := &dispatcher{
		speedProviderUpdateInterval: 1 * time.Hour,
		dispatcherUpdateInterval:    1 * time.Millisecond,
		randomSpeedProvider:         &DumbStaticSpeedProvider{speed: 10000000},
		claimers:                    make(map[IBandwidthClaimable]Weight),
		totalWeight:                 0,
		lock:                        &sync.RWMutex{},
	}
	wg := sync.WaitGroup{}
	ih1 := metainfo.NewHashFromHex("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	claimer1 := &DumbBandwidthClaimable{
		infoHash:           ih1,
		uploaded:           0,
		swarm:              &DumbSwarm{seeders: 0, leechers: 0},
		onFirstAddUploaded: func() { wg.Done() },
		addOnlyOnce:        true,
		lock:               &sync.Mutex{},
	}
	ih2 := metainfo.NewHashFromHex("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB")
	claimer2 := &DumbBandwidthClaimable{
		infoHash:           ih2,
		uploaded:           0,
		swarm:              &DumbSwarm{seeders: 100, leechers: 2500},
		onFirstAddUploaded: func() { wg.Done() },
		addOnlyOnce:        true,
		lock:               &sync.Mutex{},
	}

	wg.Add(2)
	dispatcher.ClaimOrUpdate(claimer1)
	dispatcher.ClaimOrUpdate(claimer2)

	dispatcher.Start()
	wg.Wait()
	dispatcher.Stop()
	assert.Zero(t, claimer1.uploaded)
	assert.Greater(t, claimer2.uploaded, int64(0))
}

func TestDispatcher_shouldNotRegisterIfSwarmIsNil(t *testing.T) {
	dispatcher := &dispatcher{
		claimers: make(map[IBandwidthClaimable]Weight),
		lock:     &sync.RWMutex{},
	}
	claimer1 := &DumbBandwidthClaimable{
		swarm: nil,
	}
	dispatcher.ClaimOrUpdate(claimer1)

	assert.Empty(t, dispatcher.claimers)
}
