package orchestrator

import (
	"context"
	"errors"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/utils/testutils"
	"github.com/nvn1729/congo"
	"github.com/stretchr/testify/assert"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

type mockedTierAnnouncer struct {
	lastInterval time.Duration
	annOnce      func(ctx context.Context, announce AnnouncingFunction, event tracker.AnnounceEvent) tierState
	startAnnLoop func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error)
	stopAnnLoop  func()
}

func (t *mockedTierAnnouncer) announceOnce(ctx context.Context, announce AnnouncingFunction, event tracker.AnnounceEvent) tierState {
	if t.annOnce != nil {
		return t.annOnce(ctx, announce, event)
	}
	return ALIVE
}

func (t *mockedTierAnnouncer) startAnnounceLoop(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
	if t.startAnnLoop != nil {
		return t.startAnnLoop(announce, firstEvent)
	}
	return make(chan tierState), nil
}

func (t *mockedTierAnnouncer) LastKnownInterval() time.Duration {
	return t.lastInterval
}

func (t *mockedTierAnnouncer) stopAnnounceLoop() {
	if t.stopAnnLoop != nil {
		t.stopAnnLoop()
	}
}

func Test_AllTrackersTierAnnouncer_ShouldLoopAllTrackersAndStopAllLoop(t *testing.T) {
	type startStopMockedTrackerAnnouncer struct {
		mockedTrackerAnnouncer
		wgStart *sync.WaitGroup
		wgStop  *sync.WaitGroup
	}

	var trackers []ITrackerAnnouncer

	for i := 0; i < 30; i++ {
		tr := &startStopMockedTrackerAnnouncer{
			mockedTrackerAnnouncer: mockedTrackerAnnouncer{},
			wgStart:                &sync.WaitGroup{},
			wgStop:                 &sync.WaitGroup{},
		}
		tr.startAnnLoop = func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan trackerAnnounceResult, error) {
			tr.wgStart.Done()
			return nil, nil
		}
		tr.stopAnnLoop = func() {
			tr.wgStop.Done()
		}

		tr.wgStart.Add(1)
		tr.wgStop.Add(1)

		trackers = append(trackers, tr)
	}

	tier, _ := newAllTrackersTierAnnouncer(trackers...)
	tierStates, err := tier.startAnnounceLoop(NoOpAnnouncingFun, tracker.Started)
	if err != nil {
		t.Fatal(err)
	}
	drainTierStateChanContinuously(tierStates)

	for _, tr := range trackers {
		tra := tr.(*startStopMockedTrackerAnnouncer)
		if testutils.WaitOrFailAfterTimeout(tra.wgStart, 500*time.Millisecond) != nil {
			t.Fatal("not ALL the trackers have been instruct to start")
		}
	}
	tier.stopAnnounceLoop()

	for _, tr := range trackers {
		tra := tr.(*startStopMockedTrackerAnnouncer)
		if testutils.WaitOrFailAfterTimeout(tra.wgStop, 500*time.Millisecond) != nil {
			t.Fatal("not ALL the trackers have been instruct to start")
		}
	}
}

func Test_AllTrackersTierAnnouncer_ShouldBeReusableAfterStop(t *testing.T) {
	var trackers []ITrackerAnnouncer
	var wg sync.WaitGroup

	for i := 0; i < 30; i++ {
		t := &mockedTrackerAnnouncer{}
		t.startAnnLoop = func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan trackerAnnounceResult, error) {
			wg.Done()
			return nil, nil
		}
		trackers = append(trackers, t)
	}

	tier, _ := newAllTrackersTierAnnouncer(trackers...)

	wg.Add(len(trackers))
	tierStates, err := tier.startAnnounceLoop(ThirtyMinutesIntervalNoOpAnnouncingFunc, tracker.Started)
	if err != nil {
		t.Fatal(err)
	}
	drainTierStateChanContinuously(tierStates)

	if testutils.WaitOrFailAfterTimeout(&wg, 5*time.Second) != nil {
		t.Fatal("not ALL the trackers have been instruct to announce")
	}
	tier.stopAnnounceLoop()

	wg.Add(len(trackers))
	tierStates, err = tier.startAnnounceLoop(ThirtyMinutesIntervalNoOpAnnouncingFunc, tracker.Started)
	if err != nil {
		t.Fatal(err)
	}
	drainTierStateChanContinuously(tierStates)

	if testutils.WaitOrFailAfterTimeout(&wg, 5*time.Second) != nil {
		t.Fatal("not ALL the trackers have been instruct to announce")
	}
	tier.stopAnnounceLoop()
}

func Test_AllTrackersTierAnnounce_ShouldReportStates(t *testing.T) {
	var trackers []ITrackerAnnouncer

	c1 := make(chan trackerAnnounceResult)
	t1 := &mockedTrackerAnnouncer{
		startAnnLoop: func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan trackerAnnounceResult, error) {
			return c1, nil
		},
	}
	trackers = append(trackers, t1)

	tier, _ := newAllTrackersTierAnnouncer(trackers...)

	tierStates, err := tier.startAnnounceLoop(ThirtyMinutesIntervalNoOpAnnouncingFunc, tracker.Started)
	if err != nil {
		t.Fatal(err)
	}
	defer tier.stopAnnounceLoop()

	select {
	case c1 <- trackerAnnounceResult{Interval: 24 * time.Hour}:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("tracker result was not read by tier")
	}

	select {
	case <-tierStates:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("no state received")
	}
}

func Test_AllTrackersTierAnnouncer_ShouldBeSafeToRunWithTremendousAmountOfTrackers(t *testing.T) {
	var trackers []ITrackerAnnouncer
	var latch *congo.CountDownLatch

	for i := 0; i < 3000; i++ {
		t := &mockedTrackerAnnouncer{}
		t.startAnnLoop = func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan trackerAnnounceResult, error) {
			c := make(chan trackerAnnounceResult)
			go func() {
				for {
					response, _ := announce(context.Background(), url.URL{}, tracker.None)
					time.Sleep(response.Interval)
					c <- trackerAnnounceResult{Interval: response.Interval, Completed: time.Now()}
				}
			}()
			return c, nil
		}
		trackers = append(trackers, t)
	}

	tier, _ := newAllTrackersTierAnnouncer(trackers...)
	var annFunc = buildAnnouncingFunc(1*time.Millisecond, func(u url.URL) { latch.CountDown() })

	latch = congo.NewCountDownLatch(uint(3 * len(trackers)))
	tierStates, err := tier.startAnnounceLoop(annFunc, tracker.Started)
	if err != nil {
		t.Fatal(err)
	}
	defer tier.stopAnnounceLoop()
	drainTierStateChanContinuously(tierStates)

	if !latch.WaitTimeout(10 * time.Second) {
		t.Fatal("not enough announce")
	}
}

func Test_AllTrackersTierAnnouncer_ShouldFailToBuildWithEmptyTrackerList(t *testing.T) {
	_, err := newAllTrackersTierAnnouncer()
	if err == nil || !strings.Contains(err.Error(), "empty tracker list") {
		t.Fatal("should have failed to build with empty tracker list")
	}
}

func Test_AllTrackersTierAnnouncer_StopShouldBeANoOpIfNotStarted(t *testing.T) {
	trackers := []ITrackerAnnouncer{
		&mockedTrackerAnnouncer{},
		&mockedTrackerAnnouncer{},
		&mockedTrackerAnnouncer{},
	}

	tier, _ := newAllTrackersTierAnnouncer(trackers...)

	done := make(chan struct{})
	go func() {
		tier.stopAnnounceLoop()
		tier.stopAnnounceLoop()
		tier.stopAnnounceLoop()
		done <- struct{}{}
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
	}
}

func Test_AllTrackersTierAnnouncer_ShouldAnnounceOnceToAllTrackerAndReportAliveAllSucceed(t *testing.T) {
	t1 := &mockedTrackerAnnouncer{}
	t2 := &mockedTrackerAnnouncer{}
	t3 := &mockedTrackerAnnouncer{}
	trackers := []ITrackerAnnouncer{t1, t2, t3}

	tier, _ := newAllTrackersTierAnnouncer(trackers...)

	t1.annOnce = func(ctx context.Context, announce AnnouncingFunction, event tracker.AnnounceEvent) trackerAnnounceResult {
		return trackerAnnounceResult{Err: nil}
	}
	t2.annOnce = func(ctx context.Context, announce AnnouncingFunction, event tracker.AnnounceEvent) trackerAnnounceResult {
		return trackerAnnounceResult{Err: nil}
	}
	t3.annOnce = func(ctx context.Context, announce AnnouncingFunction, event tracker.AnnounceEvent) trackerAnnounceResult {
		return trackerAnnounceResult{Err: nil}
	}

	state := tier.announceOnce(context.Background(), nil, tracker.Started)
	if state != ALIVE {
		t.Fatal("should have returned tier alive")
	}
}

func Test_AllTrackersTierAnnouncer_ShouldAnnounceOnceToAllTrackerAndReportAliveIfSomeSucceed(t *testing.T) {
	t1 := &mockedTrackerAnnouncer{}
	t2 := &mockedTrackerAnnouncer{}
	t3 := &mockedTrackerAnnouncer{}
	trackers := []ITrackerAnnouncer{t1, t2, t3}

	tier, _ := newAllTrackersTierAnnouncer(trackers...)

	t1.annOnce = func(ctx context.Context, announce AnnouncingFunction, event tracker.AnnounceEvent) trackerAnnounceResult {
		return trackerAnnounceResult{Err: nil}
	}
	t2.annOnce = func(ctx context.Context, announce AnnouncingFunction, event tracker.AnnounceEvent) trackerAnnounceResult {
		return trackerAnnounceResult{Err: errors.New("nop")}
	}
	t3.annOnce = func(ctx context.Context, announce AnnouncingFunction, event tracker.AnnounceEvent) trackerAnnounceResult {
		return trackerAnnounceResult{Err: nil}
	}

	state := tier.announceOnce(context.Background(), nil, tracker.Started)
	if state != ALIVE {
		t.Fatal("should have returned tier alive")
	}
}

func Test_AllTrackersTierAnnouncer_ShouldAnnounceOnceToAllTrackerAndReportAliveIfOneSucceed(t *testing.T) {
	t1 := &mockedTrackerAnnouncer{}
	t2 := &mockedTrackerAnnouncer{}
	t3 := &mockedTrackerAnnouncer{}
	trackers := []ITrackerAnnouncer{t1, t2, t3}

	tier, _ := newAllTrackersTierAnnouncer(trackers...)

	t1.annOnce = func(ctx context.Context, announce AnnouncingFunction, event tracker.AnnounceEvent) trackerAnnounceResult {
		return trackerAnnounceResult{Err: errors.New("nop")}
	}
	t2.annOnce = func(ctx context.Context, announce AnnouncingFunction, event tracker.AnnounceEvent) trackerAnnounceResult {
		return trackerAnnounceResult{Err: errors.New("nop")}
	}
	t3.annOnce = func(ctx context.Context, announce AnnouncingFunction, event tracker.AnnounceEvent) trackerAnnounceResult {
		return trackerAnnounceResult{Err: nil}
	}

	state := tier.announceOnce(context.Background(), nil, tracker.Started)
	if state != ALIVE {
		t.Fatal("should have returned tier alive")
	}
}

func Test_AllTrackersTierAnnouncer_ShouldAnnounceOnceToAllTrackerAndReportDeadIfAllFails(t *testing.T) {
	t1 := &mockedTrackerAnnouncer{}
	t2 := &mockedTrackerAnnouncer{}
	t3 := &mockedTrackerAnnouncer{}
	trackers := []ITrackerAnnouncer{t1, t2, t3}

	tier, _ := newAllTrackersTierAnnouncer(trackers...)

	t1.annOnce = func(ctx context.Context, announce AnnouncingFunction, event tracker.AnnounceEvent) trackerAnnounceResult {
		return trackerAnnounceResult{Err: errors.New("nop")}
	}
	t2.annOnce = func(ctx context.Context, announce AnnouncingFunction, event tracker.AnnounceEvent) trackerAnnounceResult {
		return trackerAnnounceResult{Err: errors.New("nop")}
	}
	t3.annOnce = func(ctx context.Context, announce AnnouncingFunction, event tracker.AnnounceEvent) trackerAnnounceResult {
		return trackerAnnounceResult{Err: errors.New("nop")}
	}

	state := tier.announceOnce(context.Background(), nil, tracker.Started)
	if state != DEAD {
		t.Fatal("should have returned tier dead")
	}
}

func Test_FallbackTrackersTierAnnouncer_ShouldLoopTrackersAndStopLoop(t *testing.T) {
	var trackers []ITrackerAnnouncer
	t1 := &mockedTrackerAnnouncer{}
	trackers = append(trackers, t1)
	wgStart := &sync.WaitGroup{}
	wgStop := &sync.WaitGroup{}

	wgStart.Add(1)
	wgStop.Add(1)
	t1.startAnnLoop = func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan trackerAnnounceResult, error) {
		wgStart.Done()
		return nil, nil
	}
	t1.stopAnnLoop = func() {
		wgStop.Done()
	}

	tier, _ := newFallbackTrackersTierAnnouncer(trackers...)

	tierStates, err := tier.startAnnounceLoop(ThirtyMinutesIntervalNoOpAnnouncingFunc, tracker.Started)
	if err != nil {
		t.Fatal(err)
	}
	drainTierStateChanContinuously(tierStates)

	if err := testutils.WaitOrFailAfterTimeout(wgStart, 5*time.Second); err != nil {
		t.Fatal("not started")
	}
	tier.stopAnnounceLoop()
	if err := testutils.WaitOrFailAfterTimeout(wgStop, 5*time.Second); err != nil {
		t.Fatal("not stopped")
	}
}

func Test_FallbackTrackersTierAnnouncer_ShouldBeReusableAfterStop(t *testing.T) {
	type startStopMockedTrackerAnnouncer struct {
		mockedTrackerAnnouncer
		wgStart *sync.WaitGroup
		wgStop  *sync.WaitGroup
	}

	var trackers []ITrackerAnnouncer
	t1 := &startStopMockedTrackerAnnouncer{
		mockedTrackerAnnouncer: mockedTrackerAnnouncer{},
		wgStart:                &sync.WaitGroup{},
		wgStop:                 &sync.WaitGroup{},
	}
	trackers = append(trackers, t1)

	t1.startAnnLoop = func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan trackerAnnounceResult, error) {
		t1.wgStart.Done()
		return nil, nil
	}
	t1.stopAnnLoop = func() {
		t1.wgStop.Done()
	}

	tier, _ := newFallbackTrackersTierAnnouncer(trackers...)

	t1.wgStart.Add(1)
	tierStates, err := tier.startAnnounceLoop(ThirtyMinutesIntervalNoOpAnnouncingFunc, tracker.Started)
	if err != nil {
		t.Fatal(err)
	}
	drainTierStateChanContinuously(tierStates)
	if err := testutils.WaitOrFailAfterTimeout(t1.wgStart, 500*time.Millisecond); err != nil {
		t.Fatal("not started")
	}

	t1.wgStop.Add(1)
	tier.stopAnnounceLoop()
	if err := testutils.WaitOrFailAfterTimeout(t1.wgStart, 500*time.Millisecond); err != nil {
		t.Fatal("not stopped")
	}

	t1.wgStart.Add(1)
	tierStates, err = tier.startAnnounceLoop(ThirtyMinutesIntervalNoOpAnnouncingFunc, tracker.Started)
	if err != nil {
		t.Fatal(err)
	}
	drainTierStateChanContinuously(tierStates)
	if err := testutils.WaitOrFailAfterTimeout(t1.wgStart, 500*time.Millisecond); err != nil {
		t.Fatal("not started")
	}

	t1.wgStop.Add(1)
	tier.stopAnnounceLoop()
	if err := testutils.WaitOrFailAfterTimeout(t1.wgStart, 500*time.Millisecond); err != nil {
		t.Fatal("not stopped")
	}
}

func Test_FallbackTrackersTierAnnouncer_ShouldBeSafeToRunWithTremendousAmountOfTrackers(t *testing.T) {
	var trackers []ITrackerAnnouncer
	var latch *congo.CountDownLatch

	for i := 0; i < 3000; i++ {
		t := &mockedTrackerAnnouncer{}
		t.startAnnLoop = func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan trackerAnnounceResult, error) {
			c := make(chan trackerAnnounceResult)
			go func() {
				for {
					response, _ := announce(context.Background(), url.URL{}, tracker.None)
					c <- trackerAnnounceResult{Interval: response.Interval, Completed: time.Now()}
				}
			}()
			return c, nil
		}
		trackers = append(trackers, t)
	}

	tier, _ := newFallbackTrackersTierAnnouncer(trackers...)
	var annFunc = buildAnnouncingFunc(1*time.Millisecond, func(u url.URL) { latch.CountDown() })

	latch = congo.NewCountDownLatch(uint(3 * len(trackers)))
	tierStates, err := tier.startAnnounceLoop(annFunc, tracker.Started)
	if err != nil {
		t.Fatal(err)
	}
	defer tier.stopAnnounceLoop()
	drainTierStateChanContinuously(tierStates)

	if !latch.WaitTimeout(10 * time.Second) {
		t.Fatal("not enough announce")
	}
}

func Test_FallbackTrackersTierAnnouncer_ShouldFailToBuildWithEmptyTrackerList(t *testing.T) {
	_, err := newFallbackTrackersTierAnnouncer()
	if err == nil {
		t.Fatal("Should have failed to build")
	}
}

func Test_FallbackTrackersTierAnnouncer_StopShouldBeANoOpIfNotStarted(t *testing.T) {
	trackers := []ITrackerAnnouncer{
		&mockedTrackerAnnouncer{},
		&mockedTrackerAnnouncer{},
		&mockedTrackerAnnouncer{},
	}

	tier, _ := newFallbackTrackersTierAnnouncer(trackers...)

	done := make(chan struct{})
	go func() {
		tier.stopAnnounceLoop()
		tier.stopAnnounceLoop()
		tier.stopAnnounceLoop()
		done <- struct{}{}
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
	}
}

func Test_FallbackTrackersTierAnnouncer_ShouldCallTrackerOneByOneTillOneSucceed(t *testing.T) {
	var trackers []ITrackerAnnouncer

	c1 := make(chan trackerAnnounceResult)
	t1 := &mockedTrackerAnnouncer{
		startAnnLoop: func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan trackerAnnounceResult, error) {
			return c1, nil
		},
	}
	trackers = append(trackers, t1)

	c2 := make(chan trackerAnnounceResult)
	t2 := &mockedTrackerAnnouncer{
		startAnnLoop: func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan trackerAnnounceResult, error) {
			return c2, nil
		},
	}
	trackers = append(trackers, t2)

	c3 := make(chan trackerAnnounceResult)
	t3 := &mockedTrackerAnnouncer{
		startAnnLoop: func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan trackerAnnounceResult, error) {
			return c3, nil
		},
	}
	trackers = append(trackers, t3)

	tier, _ := newFallbackTrackersTierAnnouncer(trackers...)
	tierStates, err := tier.startAnnounceLoop(ThirtyMinutesIntervalNoOpAnnouncingFunc, tracker.Started)
	if err != nil {
		t.Fatal(err)
	}
	defer tier.stopAnnounceLoop()
	drainTierStateChanContinuously(tierStates)

	select {
	case c1 <- trackerAnnounceResult{Err: errors.New("nop"), Completed: time.Now()}:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout")
	}

	select {
	case c2 <- trackerAnnounceResult{Err: errors.New("nop"), Completed: time.Now()}:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout")
	}

	select {
	case c3 <- trackerAnnounceResult{Interval: 30 * time.Minute, Completed: time.Now()}:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout")
	}
}

func Test_FallbackTrackersTierAnnouncer_ShouldStopTrackerBeforeMovingToNext(t *testing.T) {
	var trackers []ITrackerAnnouncer
	stopped := make(chan struct{})

	c1 := make(chan trackerAnnounceResult)
	t1 := &mockedTrackerAnnouncer{
		startAnnLoop: func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan trackerAnnounceResult, error) {
			return c1, nil
		},
		stopAnnLoop: func() {
			stopped <- struct{}{}
		},
	}
	trackers = append(trackers, t1, &mockedTrackerAnnouncer{})

	tier, _ := newFallbackTrackersTierAnnouncer(trackers...)
	tierStates, err := tier.startAnnounceLoop(ThirtyMinutesIntervalNoOpAnnouncingFunc, tracker.Started)
	if err != nil {
		t.Fatal(err)
	}
	defer tier.stopAnnounceLoop()
	drainTierStateChanContinuously(tierStates)

	c1 <- trackerAnnounceResult{Err: errors.New("nop")}

	select {
	case <-stopped:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout")
	}
}

func Test_FallbackTrackersTierAnnouncer_ShouldReorderTrackerListOnAnnounceSuccess(t *testing.T) {
	var trackers []ITrackerAnnouncer
	c1 := make(chan trackerAnnounceResult)
	t1 := &mockedTrackerAnnouncer{}
	t1.startAnnLoop = func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan trackerAnnounceResult, error) {
		return c1, nil
	}
	trackers = append(trackers, t1)

	c2 := make(chan trackerAnnounceResult)
	t2 := &mockedTrackerAnnouncer{}
	t2.startAnnLoop = func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan trackerAnnounceResult, error) {
		return c2, nil
	}
	trackers = append(trackers, t2)

	c3 := make(chan trackerAnnounceResult)
	t3 := &mockedTrackerAnnouncer{}
	t3.startAnnLoop = func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan trackerAnnounceResult, error) {
		return c3, nil
	}
	trackers = append(trackers, t3)

	tier, _ := newFallbackTrackersTierAnnouncer(trackers...)
	tierStates, err := tier.startAnnounceLoop(ThirtyMinutesIntervalNoOpAnnouncingFunc, tracker.Started)
	if err != nil {
		t.Fatal(err)
	}
	defer tier.stopAnnounceLoop()
	drainTierStateChanContinuously(tierStates)

	select {
	case c1 <- trackerAnnounceResult{Err: errors.New("nop"), Completed: time.Now()}:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout")
	}

	select {
	case c2 <- trackerAnnounceResult{Err: errors.New("nop"), Completed: time.Now()}:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout")
	}

	select {
	case c3 <- trackerAnnounceResult{Interval: 30 * time.Minute, Completed: time.Now()}:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout")
	}

	fallbackTier := tier.(*FallbackTrackersTierAnnouncer)
	assert.EqualValues(t, fallbackTier.tracker.list[fallbackTier.tracker.currentIndex], t3)
}

func drainTierStateChanContinuously(c <-chan tierState) {
	go func() {
		for {
			select {
			case <-c:
			}
		}
	}()
}
