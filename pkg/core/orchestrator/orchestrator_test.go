package orchestrator

import (
	"context"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/core/announcer"
	"github.com/anthonyraymond/joal-cli/pkg/utils/testutils"
	"github.com/nvn1729/congo"
	"github.com/stretchr/testify/assert"
	"net/url"
	"runtime"
	"sync"
	"testing"
	"time"
)

//noinspection GoVarAndConstTypeMayBeOmitted
var NoOpAnnouncingFun AnnouncingFunction = func(ctx context.Context, u url.URL, event tracker.AnnounceEvent) (announcer.AnnounceResponse, error) {
	return announcer.AnnounceResponse{}, nil
}
var ThirtyMinutesIntervalNoOpAnnouncingFunc = buildAnnouncingFunc(30 * time.Minute)

func buildAnnouncingFunc(interval time.Duration, callbacks ...func(u url.URL)) AnnouncingFunction {
	return func(ctx context.Context, u url.URL, event tracker.AnnounceEvent) (announcer.AnnounceResponse, error) {
		for _, c := range callbacks {
			c(u)
		}
		return announcer.AnnounceResponse{
			Interval: interval,
			Leechers: 0,
			Seeders:  0,
			Peers:    []tracker.Peer{},
		}, nil
	}
}

type dumbConfig struct {
	doesSupportAnnounceList           bool
	shouldAnnounceToAllTiers          bool
	shouldAnnounceToAllTrackersInTier bool
}

func (d *dumbConfig) DoesSupportAnnounceList() bool {
	return d.doesSupportAnnounceList
}

func (d *dumbConfig) ShouldAnnounceToAllTiers() bool {
	return d.shouldAnnounceToAllTiers
}

func (d *dumbConfig) ShouldAnnounceToAllTrackersInTier() bool {
	return d.shouldAnnounceToAllTrackersInTier
}

func Test_OrchestratorShouldFilterEmptyUrl(t *testing.T) {
	config := &dumbConfig{
		doesSupportAnnounceList:           true,
		shouldAnnounceToAllTiers:          true,
		shouldAnnounceToAllTrackersInTier: true,
	}

	o, err := NewOrchestrator(&TorrentInfo{
		Announce: "http://localhost:8000/announce",
		AnnounceList: metainfo.AnnounceList{
			{"", " ", "http://localhost:8080"},
			{"", " ", "http://localhost:9090"},
		},
	}, config)
	if err != nil {
		t.Fatal(err)
	}

	orchestrator := o.(*AllOrchestrator)

	assert.Len(t, orchestrator.tiers, 2)
	assert.Len(t, orchestrator.tiers[0].(*AllTrackersTierAnnouncer).trackers, 1)
	assert.Equal(t, orchestrator.tiers[0].(*AllTrackersTierAnnouncer).trackers[0].(*trackerAnnouncer).url.String(), "http://localhost:8080")
	assert.Len(t, orchestrator.tiers[1].(*AllTrackersTierAnnouncer).trackers, 1)
	assert.Equal(t, orchestrator.tiers[1].(*AllTrackersTierAnnouncer).trackers[0].(*trackerAnnouncer).url.String(), "http://localhost:9090")
}

func Test_FallbackOrchestrator_ShouldNotBuildWithEmptyTierList(t *testing.T) {
	_, err := newFallBackOrchestrator()
	if err == nil {
		t.Fatal("should have failed to build")
	}
}

func Test_FallbackOrchestrator_ShouldAnnounceOnlyOnFirstTierIfItSucceed(t *testing.T) {
	c1 := make(chan tierState)
	t1 := &mockedTierAnnouncer{
		startAnnLoop: func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
			return c1, nil
		},
	}

	t2HasBeenStarted := make(chan struct{})
	c2 := make(chan tierState)
	t2 := &mockedTierAnnouncer{
		startAnnLoop: func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
			t2HasBeenStarted <- struct{}{}
			return c2, nil
		},
	}

	tiers := []ITierAnnouncer{t1, t2}

	o, _ := newFallBackOrchestrator(tiers...)
	o.Start(nil)
	defer o.Stop(context.Background(), NoOpAnnouncingFun)

	c1 <- ALIVE

	runtime.Gosched()

	select {
	case <-time.After(50 * time.Millisecond):
	case <-t2HasBeenStarted:
		t.Fatal("tier2 was started")
	}

}

func Test_FallbackOrchestrator_ShouldTryTiersOneByOneUntilOneSucceed(t *testing.T) {
	c1 := make(chan tierState)
	t1 := &mockedTierAnnouncer{
		startAnnLoop: func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
			return c1, nil
		},
	}

	c2 := make(chan tierState)
	t2 := &mockedTierAnnouncer{
		startAnnLoop: func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
			return c2, nil
		},
	}

	c3 := make(chan tierState)
	t3 := &mockedTierAnnouncer{
		startAnnLoop: func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
			return c3, nil
		},
	}

	tiers := []ITierAnnouncer{t1, t2, t3}

	o, _ := newFallBackOrchestrator(tiers...)
	o.Start(nil)
	defer o.Stop(context.Background(), NoOpAnnouncingFun)

	select {
	case c1 <- DEAD:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: orchestrator did not read the event => most likely because tier has not been started by orchestrator")
	}
	select {
	case c2 <- DEAD:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: orchestrator did not read the event => most likely because tier has not been started by orchestrator")
	}
	select {
	case c3 <- ALIVE:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: orchestrator did not read the event => most likely because tier has not been started by orchestrator")
	}
}

func Test_FallbackOrchestrator_ShouldPauseBeforeReAnnouncingIfAllTiersFails(t *testing.T) {
	c1 := make(chan tierState)
	t1 := &mockedTierAnnouncer{
		startAnnLoop: func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
			return c1, nil
		},
		lastInterval: 24 * time.Hour,
	}

	c2 := make(chan tierState)
	t2 := &mockedTierAnnouncer{
		startAnnLoop: func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
			return c2, nil
		},
	}

	tiers := []ITierAnnouncer{t1, t2}

	o, _ := newFallBackOrchestrator(tiers...)
	o.Start(nil)
	defer o.Stop(context.Background(), NoOpAnnouncingFun)

	select {
	case c1 <- DEAD:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: orchestrator did not read the event => most likely because tier has not been started by orchestrator")
	}
	select {
	case c2 <- DEAD:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: orchestrator did not read the event => most likely because tier has not been started by orchestrator")
	}
	select {
	case <-time.After(50 * time.Millisecond):
	case c1 <- DEAD:
		t.Fatal("orchestrator has not wait after all tracker failed to announce")
	}
}

func Test_FallbackOrchestrator_ShouldGoBackToFirstTierIfAllFails(t *testing.T) {
	c1 := make(chan tierState)
	t1 := &mockedTierAnnouncer{
		startAnnLoop: func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
			return c1, nil
		},
		lastInterval: 0 * time.Millisecond,
	}

	c2 := make(chan tierState)
	t2 := &mockedTierAnnouncer{
		startAnnLoop: func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
			return c2, nil
		},
	}

	tiers := []ITierAnnouncer{t1, t2}

	o, _ := newFallBackOrchestrator(tiers...)
	o.Start(nil)
	defer o.Stop(context.Background(), NoOpAnnouncingFun)

	select {
	case c1 <- DEAD:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: orchestrator did not read the event => most likely because tier has not been started by orchestrator")
	}
	select {
	case c2 <- DEAD:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: orchestrator did not read the event => most likely because tier has not been started by orchestrator")
	}
	select {
	case c1 <- ALIVE:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: orchestrator did not read the event => most likely because tier has not been started by orchestrator")
	}
}

func Test_FallbackOrchestrator_ShouldReAnnounceOnFirstTrackerAfterABackupTierHasSucceed(t *testing.T) {
	c1 := make(chan tierState)
	t1 := &mockedTierAnnouncer{
		startAnnLoop: func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
			return c1, nil
		},
	}

	c2 := make(chan tierState)
	t2 := &mockedTierAnnouncer{
		startAnnLoop: func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
			return c2, nil
		},
		lastInterval: 0 * time.Millisecond,
	}

	c3 := make(chan tierState)
	t3 := &mockedTierAnnouncer{
		startAnnLoop: func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
			return c3, nil
		},
	}

	tiers := []ITierAnnouncer{t1, t2, t3}

	o, _ := newFallBackOrchestrator(tiers...)
	o.Start(nil)
	defer o.Stop(context.Background(), NoOpAnnouncingFun)

	select {
	case c1 <- DEAD:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: orchestrator did not read the event => most likely because tier has not been started by orchestrator")
	}
	select {
	case c2 <- ALIVE:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: orchestrator did not read the event => most likely because tier has not been started by orchestrator")
	}
	select {
	case c1 <- ALIVE:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: orchestrator did not read the event => most likely because tier has not been started by orchestrator")
	}

}

func Test_FallbackOrchestrator_ShouldKeepAnnouncingToFirstTrackerIfItSucceed(t *testing.T) {
	c1 := make(chan tierState)
	t1 := &mockedTierAnnouncer{
		startAnnLoop: func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
			return c1, nil
		},
		lastInterval: 0 * time.Millisecond,
	}

	c2 := make(chan tierState)
	t2 := &mockedTierAnnouncer{
		startAnnLoop: func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
			return c2, nil
		},
	}

	tiers := []ITierAnnouncer{t1, t2}

	o, _ := newFallBackOrchestrator(tiers...)
	o.Start(nil)
	defer o.Stop(context.Background(), NoOpAnnouncingFun)

	select {
	case c1 <- ALIVE:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: orchestrator did not read the event => most likely because tier has not been started by orchestrator")
	}
	select {
	case c1 <- ALIVE:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: orchestrator did not read the event => most likely because tier has not been started by orchestrator")
	}
}

func Test_FallbackOrchestrator_ShouldStopPreviousTierWhenMovingToNext(t *testing.T) {
	stoppingOne := make(chan struct{})
	c1 := make(chan tierState)
	t1 := &mockedTierAnnouncer{
		startAnnLoop: func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
			return c1, nil
		},
		stopAnnLoop: func() {
			stoppingOne <- struct{}{}
		},
	}

	c2 := make(chan tierState)
	t2 := &mockedTierAnnouncer{
		startAnnLoop: func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
			return c2, nil
		},
	}

	tiers := []ITierAnnouncer{t1, t2}

	o, _ := newFallBackOrchestrator(tiers...)
	o.Start(nil)
	defer o.Stop(context.Background(), NoOpAnnouncingFun)

	select {
	case c1 <- DEAD:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: orchestrator did not read the event => most likely because tier has not been started by orchestrator")
	}

	select {
	case <-stoppingOne:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: orchestrator did not stopped tier")
	}
}

func Test_FallbackOrchestrator_ShouldNotBlockIfStopIsCalledWhenNotStarted(t *testing.T) {
	o, _ := newFallBackOrchestrator(&mockedTierAnnouncer{}, &mockedTierAnnouncer{}, &mockedTierAnnouncer{})
	o.Start(nil)
	defer o.Stop(context.Background(), NoOpAnnouncingFun)

	done := make(chan struct{})
	go func() {
		o.Stop(context.Background(), NoOpAnnouncingFun)
		o.Stop(context.Background(), NoOpAnnouncingFun)
		o.Stop(context.Background(), NoOpAnnouncingFun)
		done <- struct{}{}
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
	}
}

func Test_FallbackOrchestrator_ShouldBeSafeToRunWithTremendousAmountOfTiers(t *testing.T) {
	var tiers []ITierAnnouncer
	var latch *congo.CountDownLatch

	for i := 0; i < 3000; i++ {
		t := &mockedTierAnnouncer{
			lastInterval: 0 * time.Millisecond,
			startAnnLoop: func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
				c := make(chan tierState)
				go func() {
					for {
						_ = latch.CountDown()
						c <- DEAD
					}
				}()
				return c, nil
			},
		}
		tiers = append(tiers, t)
	}

	latch = congo.NewCountDownLatch(uint(3 * len(tiers)))
	o, _ := newFallBackOrchestrator(tiers...)
	o.Start(nil)
	defer o.Stop(context.Background(), NoOpAnnouncingFun)

	if !latch.WaitTimeout(50 * time.Second) {
		t.Fatal("latch has not been released")
	}
}

func Test_FallbackOrchestrator_ShouldBeReusableAfterStop(t *testing.T) {
	type startStopMockedTierAnnouncer struct {
		*mockedTierAnnouncer
		wgStart *sync.WaitGroup
		wgStop  *sync.WaitGroup
	}

	c1 := make(chan tierState)
	t1 := &startStopMockedTierAnnouncer{
		mockedTierAnnouncer: &mockedTierAnnouncer{},
		wgStart:             &sync.WaitGroup{},
		wgStop:              &sync.WaitGroup{},
	}
	t1.startAnnLoop = func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
		t1.wgStart.Done()
		return c1, nil
	}
	t1.stopAnnLoop = func() {
		t1.wgStop.Done()
	}

	tiers := []ITierAnnouncer{t1}

	o, _ := newFallBackOrchestrator(tiers...)

	t1.wgStart.Add(1)
	o.Start(NoOpAnnouncingFun)
	if err := testutils.WaitOrFailAfterTimeout(t1.wgStart, 500*time.Millisecond); err != nil {
		t.Fatal("not started")
	}

	t1.wgStop.Add(1)
	o.Stop(context.Background(), NoOpAnnouncingFun)
	if err := testutils.WaitOrFailAfterTimeout(t1.wgStart, 500*time.Millisecond); err != nil {
		t.Fatal("not stopped")
	}

	t1.wgStart.Add(1)
	o.Start(NoOpAnnouncingFun)
	if err := testutils.WaitOrFailAfterTimeout(t1.wgStart, 500*time.Millisecond); err != nil {
		t.Fatal("not started")
	}

	t1.wgStop.Add(1)
	o.Stop(context.Background(), NoOpAnnouncingFun)
	if err := testutils.WaitOrFailAfterTimeout(t1.wgStart, 500*time.Millisecond); err != nil {
		t.Fatal("not stopped")
	}
}

func Test_FallbackOrchestrator_ShouldAnnounceStopOnStop(t *testing.T) {
	callAnnounceStop := make(chan struct{})
	c1 := make(chan tierState)
	t1 := &mockedTierAnnouncer{
		startAnnLoop: func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
			return c1, nil
		},
		annOnce: func(ctx context.Context, announce AnnouncingFunction, event tracker.AnnounceEvent) tierState {
			if event == tracker.Stopped {
				close(callAnnounceStop)
			}
			return ALIVE
		},
	}

	c2 := make(chan tierState)
	t2 := &mockedTierAnnouncer{
		startAnnLoop: func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
			return c2, nil
		},
	}

	tiers := []ITierAnnouncer{t1, t2}

	o, _ := newFallBackOrchestrator(tiers...)
	o.Start(nil)

	select {
	case c1 <- ALIVE:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: orchestrator did not read the event => most likely because tier has not been started by orchestrator")
	}

	go o.Stop(context.Background(), NoOpAnnouncingFun)

	select {
	case <-callAnnounceStop:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: orchestrator did not sent STOP announce")
	}
}

func Test_FallbackOrchestrator_ShouldAnnounceStopAndReturnIfAllTiersFails(t *testing.T) {
	t1STOP := make(chan struct{})
	c1 := make(chan tierState)
	t1 := &mockedTierAnnouncer{
		startAnnLoop: func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
			return c1, nil
		},
		annOnce: func(ctx context.Context, announce AnnouncingFunction, event tracker.AnnounceEvent) tierState {
			if event == tracker.Stopped {
				t1STOP <- struct{}{}
			}
			return DEAD
		},
	}

	t2STOP := make(chan struct{})
	c2 := make(chan tierState)
	t2 := &mockedTierAnnouncer{
		startAnnLoop: func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
			return c2, nil
		},
		annOnce: func(ctx context.Context, announce AnnouncingFunction, event tracker.AnnounceEvent) tierState {
			if event == tracker.Stopped {
				t2STOP <- struct{}{}
			}
			return DEAD
		},
	}

	tiers := []ITierAnnouncer{t1, t2}

	o, _ := newFallBackOrchestrator(tiers...)
	o.Start(nil)

	select {
	case c1 <- ALIVE:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: orchestrator did not read the event => most likely because tier has not been started by orchestrator")
	}

	doneStop := make(chan tierState)
	go func() {
		o.Stop(context.Background(), NoOpAnnouncingFun)
		close(doneStop)
	}()

	select {
	case <-t1STOP:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: orchestrator did not sent STOP announce")
	}

	select {
	case <-doneStop:
	case <-time.After(5 * time.Second):
		t.Fatal("stop has not ended")
	}
}

func Test_AllOrchestrator_ShouldNotBuildWithEmptyTierList(t *testing.T) {
	_, err := newAllOrchestrator()
	if err == nil {
		t.Fatal("should have failed to build")
	}
}

func Test_AllOrchestrator_ShouldAnnounceOnAllTiers(t *testing.T) {
	c1 := make(chan tierState)
	t1 := &mockedTierAnnouncer{
		startAnnLoop: func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
			return c1, nil
		},
	}

	c2 := make(chan tierState)
	t2 := &mockedTierAnnouncer{
		startAnnLoop: func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
			return c2, nil
		},
	}

	tiers := []ITierAnnouncer{t1, t2}

	o, _ := newAllOrchestrator(tiers...)
	o.Start(nil)
	defer o.Stop(context.Background(), NoOpAnnouncingFun)

	select {
	case c1 <- ALIVE:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: orchestrator did not read the event => most likely because tier has not been started by orchestrator")
	}
	select {
	case c2 <- ALIVE:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: orchestrator did not read the event => most likely because tier has not been started by orchestrator")
	}
}

func Test_AllOrchestrator_ShouldContinueAnnouncingEvenIfAllTierFails(t *testing.T) {
	c1 := make(chan tierState)
	t1 := &mockedTierAnnouncer{
		startAnnLoop: func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
			return c1, nil
		},
	}

	c2 := make(chan tierState)
	t2 := &mockedTierAnnouncer{
		startAnnLoop: func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
			return c2, nil
		},
	}

	tiers := []ITierAnnouncer{t1, t2}

	o, _ := newAllOrchestrator(tiers...)
	o.Start(nil)
	defer o.Stop(context.Background(), NoOpAnnouncingFun)

	select {
	case c1 <- DEAD:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: orchestrator did not read the event => most likely because tier has not been started by orchestrator")
	}
	select {
	case c2 <- DEAD:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: orchestrator did not read the event => most likely because tier has not been started by orchestrator")
	}
	select {
	case c1 <- DEAD:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: orchestrator did not read the event => most likely because tier has not been started by orchestrator")
	}
	select {
	case c2 <- DEAD:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: orchestrator did not read the event => most likely because tier has not been started by orchestrator")
	}
}

func Test_AllOrchestrator_ShouldStartAndStopLoop(t *testing.T) {
	type startStopMockedTierAnnouncer struct {
		*mockedTierAnnouncer
		wgStart *sync.WaitGroup
		wgStop  *sync.WaitGroup
	}

	var tiers []ITierAnnouncer

	for i := 0; i < 30; i++ {
		tier := &startStopMockedTierAnnouncer{
			mockedTierAnnouncer: &mockedTierAnnouncer{},
			wgStart:             &sync.WaitGroup{},
			wgStop:              &sync.WaitGroup{},
		}
		tier.startAnnLoop = func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
			tier.wgStart.Done()
			return make(chan tierState), nil
		}
		tier.stopAnnLoop = func() {
			tier.wgStop.Done()
		}

		tier.wgStart.Add(1)
		tier.wgStop.Add(1)

		tiers = append(tiers, tier)
	}

	o, _ := newAllOrchestrator(tiers...)
	o.Start(nil)

	for _, ti := range tiers {
		tier := ti.(*startStopMockedTierAnnouncer)
		if testutils.WaitOrFailAfterTimeout(tier.wgStart, 500*time.Millisecond) != nil {
			t.Fatal("not ALL the trackers have been instruct to start")
		}
	}

	o.Stop(context.Background(), NoOpAnnouncingFun)
	for _, ti := range tiers {
		tier := ti.(*startStopMockedTierAnnouncer)
		if testutils.WaitOrFailAfterTimeout(tier.wgStop, 500*time.Millisecond) != nil {
			t.Fatal("not ALL the trackers have been instruct to start")
		}
	}
}

func Test_AllOrchestrator_ShouldNotBlockIfStopIsCalledWhenNotStarted(t *testing.T) {
	o, _ := newAllOrchestrator(&mockedTierAnnouncer{}, &mockedTierAnnouncer{}, &mockedTierAnnouncer{})
	o.Start(nil)
	defer o.Stop(context.Background(), NoOpAnnouncingFun)

	done := make(chan struct{})
	go func() {
		o.Stop(context.Background(), NoOpAnnouncingFun)
		o.Stop(context.Background(), NoOpAnnouncingFun)
		o.Stop(context.Background(), NoOpAnnouncingFun)
		done <- struct{}{}
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
	}
}

func Test_AllOrchestrator_ShouldBeSafeToRunWithTremendousAmountOfTiers(t *testing.T) {
	var tiers []ITierAnnouncer
	var latch *congo.CountDownLatch

	for i := 0; i < 3000; i++ {
		t := &mockedTierAnnouncer{
			lastInterval: 0 * time.Millisecond,
			startAnnLoop: func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
				c := make(chan tierState)
				go func() {
					for {
						_ = latch.CountDown()
						c <- DEAD
						runtime.Gosched()
					}
				}()
				return c, nil
			},
		}
		tiers = append(tiers, t)
	}

	latch = congo.NewCountDownLatch(uint(3 * len(tiers)))
	o, _ := newAllOrchestrator(tiers...)
	o.Start(nil)
	defer o.Stop(context.Background(), NoOpAnnouncingFun)

	if !latch.WaitTimeout(50 * time.Second) {
		t.Fatal("latch has not been released")
	}
}

func Test_AllOrchestrator_ShouldBeReusableAfterStop(t *testing.T) {
	type startStopMockedTierAnnouncer struct {
		*mockedTierAnnouncer
		wgStart *sync.WaitGroup
		wgStop  *sync.WaitGroup
	}

	c1 := make(chan tierState)
	t1 := &startStopMockedTierAnnouncer{
		mockedTierAnnouncer: &mockedTierAnnouncer{},
		wgStart:             &sync.WaitGroup{},
		wgStop:              &sync.WaitGroup{},
	}
	t1.startAnnLoop = func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
		t1.wgStart.Done()
		return c1, nil
	}
	t1.stopAnnLoop = func() {
		t1.wgStop.Done()
	}

	tiers := []ITierAnnouncer{t1}

	o, _ := newAllOrchestrator(tiers...)

	t1.wgStart.Add(1)
	o.Start(NoOpAnnouncingFun)
	if err := testutils.WaitOrFailAfterTimeout(t1.wgStart, 500*time.Millisecond); err != nil {
		t.Fatal("not started")
	}

	t1.wgStop.Add(1)
	o.Stop(context.Background(), NoOpAnnouncingFun)
	if err := testutils.WaitOrFailAfterTimeout(t1.wgStart, 500*time.Millisecond); err != nil {
		t.Fatal("not stopped")
	}

	t1.wgStart.Add(1)
	o.Start(NoOpAnnouncingFun)
	if err := testutils.WaitOrFailAfterTimeout(t1.wgStart, 500*time.Millisecond); err != nil {
		t.Fatal("not started")
	}

	t1.wgStop.Add(1)
	o.Stop(context.Background(), NoOpAnnouncingFun)
	if err := testutils.WaitOrFailAfterTimeout(t1.wgStart, 500*time.Millisecond); err != nil {
		t.Fatal("not stopped")
	}
}

func Test_AllOrchestrator_ShouldAnnounceStopOnStop(t *testing.T) {
	t1STOP := make(chan struct{}, 1)
	c1 := make(chan tierState)
	t1 := &mockedTierAnnouncer{
		startAnnLoop: func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
			return c1, nil
		},
		annOnce: func(ctx context.Context, announce AnnouncingFunction, event tracker.AnnounceEvent) tierState {
			if event == tracker.Stopped {
				t1STOP <- struct{}{}
			}
			return ALIVE
		},
	}

	t2STOP := make(chan struct{}, 1)
	c2 := make(chan tierState)
	t2 := &mockedTierAnnouncer{
		startAnnLoop: func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
			return c2, nil
		},
		annOnce: func(ctx context.Context, announce AnnouncingFunction, event tracker.AnnounceEvent) tierState {
			if event == tracker.Stopped {
				t2STOP <- struct{}{}
			}
			return DEAD
		},
	}

	tiers := []ITierAnnouncer{t1, t2}

	o, _ := newAllOrchestrator(tiers...)
	o.Start(nil)

	select {
	case c1 <- DEAD:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: orchestrator did not read the event => most likely because tier has not been started by orchestrator")
	}

	doneStop := make(chan tierState)
	go func() {
		o.Stop(context.Background(), NoOpAnnouncingFun)
		close(doneStop)
	}()

	select {
	case <-t1STOP:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: orchestrator did not sent STOP announce")
	}
	select {
	case <-t2STOP:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: orchestrator did not sent STOP announce")
	}

	select {
	case <-doneStop:
	case <-time.After(5 * time.Second):
		t.Fatal("stop has not ended")
	}
}

func Test_AllOrchestrator_ShouldAnnounceStopOnStopAndReturnIfNoneSucceed(t *testing.T) {
	t1STOP := make(chan struct{})
	c1 := make(chan tierState)
	t1 := &mockedTierAnnouncer{
		startAnnLoop: func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
			return c1, nil
		},
		annOnce: func(ctx context.Context, announce AnnouncingFunction, event tracker.AnnounceEvent) tierState {
			if event == tracker.Stopped {
				t1STOP <- struct{}{}
			}
			return DEAD
		},
	}

	t2STOP := make(chan struct{})
	c2 := make(chan tierState)
	t2 := &mockedTierAnnouncer{
		startAnnLoop: func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
			return c2, nil
		},
		annOnce: func(ctx context.Context, announce AnnouncingFunction, event tracker.AnnounceEvent) tierState {
			if event == tracker.Stopped {
				t2STOP <- struct{}{}
			}
			return DEAD
		},
	}

	tiers := []ITierAnnouncer{t1, t2}

	o, _ := newAllOrchestrator(tiers...)
	o.Start(nil)

	select {
	case c1 <- DEAD:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: orchestrator did not read the event => most likely because tier has not been started by orchestrator")
	}

	doneStop := make(chan tierState)
	go func() {
		o.Stop(context.Background(), NoOpAnnouncingFun)
		close(doneStop)
	}()

	select {
	case <-t1STOP:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: orchestrator did not sent STOP announce")
	}
	select {
	case <-t2STOP:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: orchestrator did not sent STOP announce")
	}

	select {
	case <-doneStop:
	case <-time.After(5 * time.Second):
		t.Fatal("stop has not ended")
	}
}
