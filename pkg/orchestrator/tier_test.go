package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/utils/testutils"
	"github.com/golang/mock/gomock"
	"github.com/nvn1729/congo"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

func Test_AllTrackersTierAnnouncer_ShouldLoopAllTrackersAndStopAllLoop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var trackers []ITrackerAnnouncer
	var wg sync.WaitGroup

	for i := 0; i < 30; i++ {
		wg.Add(1)
		t := NewMockITrackerAnnouncer(ctrl)
		t.
			EXPECT().
			startAnnounceLoop(gomock.Any(), gomock.Eq(tracker.Started)).
			Do(func(a AnnouncingFunction, e tracker.AnnounceEvent) { wg.Done() }).
			Times(1)
		t.EXPECT().stopAnnounceLoop().Times(1)

		t.EXPECT().Responses().Return(make(chan trackerAnnounceResult)).AnyTimes()

		trackers = append(trackers, t)
	}

	tier, _ := newAllTrackersTierAnnouncer(trackers...)
	go tier.startAnnounceLoop(ThirtyMinutesIntervalNoOpAnnouncingFunc, tracker.Started)
	defer tier.stopAnnounceLoop()

	if testutils.WaitOrFailAfterTimeout(&wg, 50*time.Millisecond) != nil {
		t.Fatal("not ALL the trackers have been instruct to announce")
	}
}

func Test_AllTrackersTierAnnouncer_ShouldBeReusableAfterStop(t *testing.T) {
	var trackers []ITrackerAnnouncer
	var wg sync.WaitGroup

	for i := 0; i < 30; i++ {
		trackerUrl := testutils.MustParseUrl(fmt.Sprintf("http://localhost/%d", i))
		trackers = append(trackers, newTracker(*trackerUrl))
	}

	tier, _ := newAllTrackersTierAnnouncer(trackers...)
	var annFunc = buildAnnouncingFunc(1*time.Minute, func(u url.URL) { wg.Done() })

	wg.Add(len(trackers))
	go tier.startAnnounceLoop(annFunc, tracker.Started)
	if testutils.WaitOrFailAfterTimeout(&wg, 50*time.Millisecond) != nil {
		t.Fatal("not ALL the trackers have been instruct to announce")
	}
	tier.stopAnnounceLoop()

	wg.Add(len(trackers))
	go tier.startAnnounceLoop(annFunc, tracker.Started)
	if testutils.WaitOrFailAfterTimeout(&wg, 50*time.Millisecond) != nil {
		t.Fatal("not ALL the trackers have been instruct to announce")
	}
	tier.stopAnnounceLoop()
}

func Test_AllTrackersTierAnnouncer_ShouldConsiderTierDeadIfAllTrackerFails(t *testing.T) {
	var trackers []ITrackerAnnouncer

	for i := 0; i < 30; i++ {
		trackerUrl := testutils.MustParseUrl(fmt.Sprintf("http://localhost/%d", i))
		trackers = append(trackers, newTracker(*trackerUrl))
	}

	tier, _ := newAllTrackersTierAnnouncer(trackers...)

	go tier.startAnnounceLoop(ErrorAnnouncingFunc, tracker.Started)
	defer tier.stopAnnounceLoop()

	timeout := time.Now().Add(1 * time.Second)
	for {
		timeoutChan := time.After(timeout.Sub(time.Now()))

		select {
		case <-timeoutChan:
			t.Fatal("timout reached, tracker has not reported is state as DEAD")
		case s := <-tier.States():
			if s == DEAD {
				// success
				return
			}
			// if alive continue
		}
	}
}

func Test_AllTrackersTierAnnouncer_ShouldNotReportAliveAfterFirstAnnounceFailedButOtherNotAnswered(t *testing.T) {
	var trackers []ITrackerAnnouncer

	for i := 0; i < 30; i++ {
		trackerUrl := testutils.MustParseUrl(fmt.Sprintf("http://localhost/%d", i))
		trackers = append(trackers, newTracker(*trackerUrl))
	}

	tier, _ := newAllTrackersTierAnnouncer(trackers...)
	lock := &sync.Mutex{}
	latch := congo.NewCountDownLatch(1)
	var annFunc = buildErrAnnouncingFunc(func(u url.URL) {
		lock.Lock() // first call lock the mutex to prevent any other announce to run
		defer latch.CountDown()
	})

	go tier.startAnnounceLoop(annFunc, tracker.Started)
	defer tier.stopAnnounceLoop()

	if !latch.WaitTimeout(50 * time.Millisecond) {
		t.Fatalf("should have released the latch")
	}

	select {
	case <-time.After(50 * time.Millisecond):
		// perfect, it has not reported his state
	case <-tier.States():
		t.Fatalf("should not have reported his state yet")
	}
}

func Test_AllTrackersTierAnnouncer_ShouldReportAliveAfterFirstAnnounceSuccess(t *testing.T) {
	var trackers []ITrackerAnnouncer

	for i := 0; i < 30; i++ {
		trackerUrl := testutils.MustParseUrl(fmt.Sprintf("http://localhost/%d", i))
		trackers = append(trackers, newTracker(*trackerUrl))
	}

	tier, _ := newAllTrackersTierAnnouncer(trackers...)
	lock := &sync.Mutex{}
	latch := congo.NewCountDownLatch(1)
	var annFunc = buildAnnouncingFunc(1800*time.Second, func(u url.URL) {
		lock.Lock() // first call lock the mutex to prevent any other announce to run
		defer latch.CountDown()
	})

	go tier.startAnnounceLoop(annFunc, tracker.Started)
	defer tier.stopAnnounceLoop()

	if !latch.WaitTimeout(50 * time.Millisecond) {
		t.Fatalf("should have released the latch")
	}

	select {
	case <-time.After(50 * time.Millisecond):
		t.Fatalf("should have reported")
	case s := <-tier.States():
		if s != ALIVE {
			t.Fatalf("should have reported state ALIVE")
		}
	}
}

func Test_AllTrackersTierAnnouncer_ShouldConsiderTierDeadIfAllTrackerFailsWithSingleTracker(t *testing.T) {
	var trackers []ITrackerAnnouncer

	for i := 0; i < 1; i++ {
		trackerUrl := testutils.MustParseUrl(fmt.Sprintf("http://localhost/%d", i))
		trackers = append(trackers, newTracker(*trackerUrl))
	}

	tier, _ := newAllTrackersTierAnnouncer(trackers...)

	go tier.startAnnounceLoop(ErrorAnnouncingFunc, tracker.Started)
	defer tier.stopAnnounceLoop()

	timeout := time.Now().Add(1 * time.Second)
	for {
		timeoutChan := time.After(timeout.Sub(time.Now()))

		select {
		case <-timeoutChan:
			t.Fatal("timout reached, tracker has not reported is state as DEAD")
		case s := <-tier.States():
			if s == DEAD {
				// success
				return
			}
			// if alive continue
		}
	}
}

func Test_AllTrackersTierAnnouncer_ShouldReconsiderDeadTierAliveIfOneTrackerSucceed(t *testing.T) {
	var trackers []ITrackerAnnouncer

	for i := 0; i < 10; i++ {
		trackerUrl := testutils.MustParseUrl(fmt.Sprintf("http://localhost/%d", i))
		trackers = append(trackers, newTracker(*trackerUrl))
	}

	DefaultDurationWaitOnError = 1 * time.Millisecond
	announceResponse := &tracker.AnnounceResponse{}
	var errResponse = errors.New("nop")
	tier, _ := newAllTrackersTierAnnouncer(trackers...)
	//noinspection GoVarAndConstTypeMayBeOmitted
	var failAnnFunc AnnouncingFunction = func(ctx context.Context, u url.URL, event tracker.AnnounceEvent) (tracker.AnnounceResponse, error) {
		return *announceResponse, errResponse
	}

	go tier.startAnnounceLoop(failAnnFunc, tracker.Started)
	defer tier.stopAnnounceLoop()

	timeout := time.Now().Add(1 * time.Second)
	found := false
	for found != true {
		timeoutChan := time.After(timeout.Sub(time.Now()))

		select {
		case <-timeoutChan:
			t.Fatal("timout reached, tracker has not reported his state as DEAD")
		case s := <-tier.States():
			if s == DEAD {
				found = true
			}
			// if alive continue
		}
	}

	announceResponse = &tracker.AnnounceResponse{Interval: 1800}
	errResponse = nil

	timeout = time.Now().Add(1 * time.Second)
	found = false
	for found != true {
		timeoutChan := time.After(timeout.Sub(time.Now()))

		select {
		case <-timeoutChan:
			t.Fatal("timout reached, tracker has not reported his state as ALIVE")
		case s := <-tier.States():
			if s == ALIVE {
				found = true
			}
			// if alive continue
		}
	}
}

func Test_AllTrackersTierAnnouncer_ShouldNotPreventStopIfATrackerIsTakingForeverToAnnounce(t *testing.T) {
	var trackers []ITrackerAnnouncer

	for i := 0; i < 10; i++ {
		trackerUrl := testutils.MustParseUrl(fmt.Sprintf("http://localhost/%d", i))
		trackers = append(trackers, newTracker(*trackerUrl))
	}
	latch := congo.NewCountDownLatch(uint(len(trackers)))

	tier, _ := newAllTrackersTierAnnouncer(trackers...)
	var annFunc = buildAnnouncingFunc(1*time.Millisecond, func(u url.URL) {
		latch.CountDown()
		time.Sleep(500 * time.Minute)
	})

	go tier.startAnnounceLoop(annFunc, tracker.Started)

	if !latch.WaitTimeout(500 * time.Millisecond) {
		t.Fatal("latch has not been released")
	}
	time.Sleep(5 * time.Millisecond) // Allow for the time.sleep to trigger in the annonce funcs

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		tier.stopAnnounceLoop()
		wg.Done()
	}()

	if err := testutils.WaitOrFailAfterTimeout(wg, 500*time.Millisecond); err != nil {
		t.Fatal("Should have stopped before the timeout")
	}
}

func Test_AllTrackersTierAnnouncer_ShouldBeSafeToRunWithTremendousAmountOfTrackers(t *testing.T) {
	var trackers []ITrackerAnnouncer
	var latch *congo.CountDownLatch

	for i := 0; i < 3000; i++ {
		trackerUrl := testutils.MustParseUrl(fmt.Sprintf("http://localhost/%d", i))
		trackers = append(trackers, newTracker(*trackerUrl))
	}

	tier, _ := newAllTrackersTierAnnouncer(trackers...)
	var annFunc = buildAnnouncingFunc(1*time.Millisecond, func(u url.URL) { latch.CountDown() })

	latch = congo.NewCountDownLatch(uint(5 * len(trackers)))
	go tier.startAnnounceLoop(annFunc, tracker.Started)
	defer tier.stopAnnounceLoop()

	if !latch.WaitTimeout(500 * time.Millisecond) {
		t.Fatal("not enough announce")
	}
}

func Test_AllTrackersTierAnnouncer_ShouldFailToBuildWithEmptyTrackerList(t *testing.T) {
	_, err := newAllTrackersTierAnnouncer()
	if err == nil || !strings.Contains(err.Error(), "empty tracker list") {
		t.Fatal("should have failed to build with empty tracker list")
	}
}

func Test_AllTrackersTierAnnouncer_ShouldNotBlockWhenStopAnnounceLoopIsCalledButTheTierWasNotStarted(t *testing.T) {
	var trackers []ITrackerAnnouncer

	for i := 0; i < 1; i++ {
		trackerUrl := testutils.MustParseUrl(fmt.Sprintf("http://localhost/%d", i))
		trackers = append(trackers, newTracker(*trackerUrl))
	}

	tier, _ := newAllTrackersTierAnnouncer(trackers...)

	latch := congo.NewCountDownLatch(1)
	go func() {
		tier.stopAnnounceLoop()
		tier.stopAnnounceLoop()
		tier.stopAnnounceLoop()
		tier.stopAnnounceLoop()
		tier.stopAnnounceLoop()
		tier.stopAnnounceLoop()
		latch.CountDown()
	}()

	if !latch.WaitTimeout(500 * time.Millisecond) {
		t.Fatal("Should not have blocked")
	}
}

func Test_AllTrackersTierAnnouncer_ShouldAnnounceOnceToAllTrackerAndReportAliveAllSucceed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var trackers []ITrackerAnnouncer
	t1 := NewMockITrackerAnnouncer(ctrl)
	c1 := make(chan trackerAnnounceResult)
	t1.EXPECT().Responses().Return(c1).AnyTimes()
	t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).AnyTimes()
	t1.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t1)

	t2 := NewMockITrackerAnnouncer(ctrl)
	c2 := make(chan trackerAnnounceResult)
	t2.EXPECT().Responses().Return(c2).AnyTimes()
	t2.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).AnyTimes()
	t2.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t2)

	t3 := NewMockITrackerAnnouncer(ctrl)
	c3 := make(chan trackerAnnounceResult)
	t3.EXPECT().Responses().Return(c3).AnyTimes()
	t3.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).AnyTimes()
	t3.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t3)

	tier, _ := newAllTrackersTierAnnouncer(trackers...)

	t1.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).Return(trackerAnnounceResult{Err: nil}).Times(1)
	t2.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).Return(trackerAnnounceResult{Err: nil}).Times(1)
	t3.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).Return(trackerAnnounceResult{Err: nil}).Times(1)

	state := tier.announceOnce(context.Background(), nil, tracker.Started)
	if state != ALIVE {
		t.Fatal("should have returned tier alive")
	}
}

func Test_AllTrackersTierAnnouncer_ShouldAnnounceOnceToAllTrackerAndReportAliveIfSOmeSucceed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var trackers []ITrackerAnnouncer
	t1 := NewMockITrackerAnnouncer(ctrl)
	c1 := make(chan trackerAnnounceResult)
	t1.EXPECT().Responses().Return(c1).AnyTimes()
	t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).AnyTimes()
	t1.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t1)

	t2 := NewMockITrackerAnnouncer(ctrl)
	c2 := make(chan trackerAnnounceResult)
	t2.EXPECT().Responses().Return(c2).AnyTimes()
	t2.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).AnyTimes()
	t2.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t2)

	t3 := NewMockITrackerAnnouncer(ctrl)
	c3 := make(chan trackerAnnounceResult)
	t3.EXPECT().Responses().Return(c3).AnyTimes()
	t3.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).AnyTimes()
	t3.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t3)

	tier, _ := newAllTrackersTierAnnouncer(trackers...)

	t1.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).Return(trackerAnnounceResult{Err: nil}).Times(1)
	t2.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).Return(trackerAnnounceResult{Err: errors.New("nop")}).Times(1)
	t3.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).Return(trackerAnnounceResult{Err: nil}).Times(1)

	state := tier.announceOnce(context.Background(), nil, tracker.Started)
	if state != ALIVE {
		t.Fatal("should have returned tier alive")
	}
}

func Test_AllTrackersTierAnnouncer_ShouldAnnounceOnceToAllTrackerAndReportAliveIfOneSucceed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var trackers []ITrackerAnnouncer
	t1 := NewMockITrackerAnnouncer(ctrl)
	c1 := make(chan trackerAnnounceResult)
	t1.EXPECT().Responses().Return(c1).AnyTimes()
	t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).AnyTimes()
	t1.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t1)

	t2 := NewMockITrackerAnnouncer(ctrl)
	c2 := make(chan trackerAnnounceResult)
	t2.EXPECT().Responses().Return(c2).AnyTimes()
	t2.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).AnyTimes()
	t2.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t2)

	t3 := NewMockITrackerAnnouncer(ctrl)
	c3 := make(chan trackerAnnounceResult)
	t3.EXPECT().Responses().Return(c3).AnyTimes()
	t3.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).AnyTimes()
	t3.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t3)

	tier, _ := newAllTrackersTierAnnouncer(trackers...)

	t1.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).Return(trackerAnnounceResult{Err: nil}).Times(1)
	t2.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).Return(trackerAnnounceResult{Err: errors.New("nop")}).Times(1)
	t3.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).Return(trackerAnnounceResult{Err: errors.New("nop")}).Times(1)

	state := tier.announceOnce(context.Background(), nil, tracker.Started)
	if state != ALIVE {
		t.Fatal("should have returned tier alive")
	}
}

func Test_AllTrackersTierAnnouncer_ShouldAnnounceOnceToAllTrackerAndReportDeadIfAllFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var trackers []ITrackerAnnouncer
	t1 := NewMockITrackerAnnouncer(ctrl)
	c1 := make(chan trackerAnnounceResult)
	t1.EXPECT().Responses().Return(c1).AnyTimes()
	t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).AnyTimes()
	t1.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t1)

	t2 := NewMockITrackerAnnouncer(ctrl)
	c2 := make(chan trackerAnnounceResult)
	t2.EXPECT().Responses().Return(c2).AnyTimes()
	t2.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).AnyTimes()
	t2.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t2)

	t3 := NewMockITrackerAnnouncer(ctrl)
	c3 := make(chan trackerAnnounceResult)
	t3.EXPECT().Responses().Return(c3).AnyTimes()
	t3.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).AnyTimes()
	t3.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t3)

	tier, _ := newAllTrackersTierAnnouncer(trackers...)

	t1.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).Return(trackerAnnounceResult{Err: errors.New("nop")}).Times(1)
	t2.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).Return(trackerAnnounceResult{Err: errors.New("nop")}).Times(1)
	t3.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).Return(trackerAnnounceResult{Err: errors.New("nop")}).Times(1)

	state := tier.announceOnce(context.Background(), nil, tracker.Started)
	if state != DEAD {
		t.Fatal("should have returned tier dead")
	}
}

func Test_FallbackTrackersTierAnnouncer_ShouldLoopTrackersAndStopLoop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var trackers []ITrackerAnnouncer
	t1 := NewMockITrackerAnnouncer(ctrl)
	c1 := make(chan trackerAnnounceResult)
	t1.EXPECT().Responses().Return(c1).AnyTimes()
	t1.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	trackers = append(trackers, t1)

	latch := congo.NewCountDownLatch(1)
	gomock.InOrder(
		t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).Do(func(anF AnnouncingFunction, e tracker.AnnounceEvent) {
			defer latch.CountDown()
			c1 <- trackerAnnounceResult{Err: nil, Interval: 1800 * time.Second, Completed: time.Now()}
		}).Times(1),
		t1.EXPECT().stopAnnounceLoop().Times(1),
	)

	tier, _ := newFallbackTrackersTierAnnouncer(trackers...)

	go tier.startAnnounceLoop(ThirtyMinutesIntervalNoOpAnnouncingFunc, tracker.Started)

	if !latch.WaitTimeout(50 * time.Millisecond) {
		t.Fatal("latch has not released")
	}

	tier.stopAnnounceLoop()
}

func Test_FallbackTrackersTierAnnouncer_ShouldBeReusableAfterStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var trackers []ITrackerAnnouncer
	t1 := NewMockITrackerAnnouncer(ctrl)
	c1 := make(chan trackerAnnounceResult)
	t1.EXPECT().Responses().Return(c1).AnyTimes()
	t1.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	t1.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t1)

	latch := congo.NewCountDownLatch(1)
	gomock.InOrder(
		t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).Do(func(anF AnnouncingFunction, e tracker.AnnounceEvent) {
			defer latch.CountDown()
			c1 <- trackerAnnounceResult{Err: nil, Interval: 1800 * time.Second, Completed: time.Now()}
		}).Times(1),
		t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).Do(func(anF AnnouncingFunction, e tracker.AnnounceEvent) {
			defer latch.CountDown()
			c1 <- trackerAnnounceResult{Err: nil, Interval: 1800 * time.Second, Completed: time.Now()}
		}).Times(1),
	)

	tier, _ := newFallbackTrackersTierAnnouncer(trackers...)

	go tier.startAnnounceLoop(ThirtyMinutesIntervalNoOpAnnouncingFunc, tracker.Started)
	if !latch.WaitTimeout(50 * time.Millisecond) {
		t.Fatal("latch has not released")
	}
	tier.stopAnnounceLoop()

	latch = congo.NewCountDownLatch(1)
	go tier.startAnnounceLoop(ThirtyMinutesIntervalNoOpAnnouncingFunc, tracker.Started)
	if !latch.WaitTimeout(50 * time.Millisecond) {
		t.Fatal("latch has not released")
	}
	tier.stopAnnounceLoop()
}

func Test_FallbackTrackersTierAnnouncer_ShouldConsiderTierDeadIfAllTrackerFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var trackers []ITrackerAnnouncer
	t1 := NewMockITrackerAnnouncer(ctrl)
	c1 := make(chan trackerAnnounceResult)
	t1.EXPECT().Responses().Return(c1).AnyTimes()
	t1.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	t1.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t1)

	t2 := NewMockITrackerAnnouncer(ctrl)
	c2 := make(chan trackerAnnounceResult)
	t2.EXPECT().Responses().Return(c2).AnyTimes()
	t2.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	t2.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t2)

	gomock.InOrder(
		t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).Do(func(anF AnnouncingFunction, e tracker.AnnounceEvent) {
			c1 <- trackerAnnounceResult{Err: errors.New("nop")}
		}).Times(1),
		t2.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).Do(func(anF AnnouncingFunction, e tracker.AnnounceEvent) {
			c2 <- trackerAnnounceResult{Err: errors.New("nop")}
		}).Times(1),
	)

	tier, _ := newFallbackTrackersTierAnnouncer(trackers...)

	go tier.startAnnounceLoop(ThirtyMinutesIntervalNoOpAnnouncingFunc, tracker.Started)
	defer tier.stopAnnounceLoop()

	select {
	case st := <-tier.States():
		if st != DEAD {
			t.Fatalf("should have reported tier DEAD, %v received", st)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("should have reported state before timeout")
	}
}

func Test_FallbackTrackersTierAnnouncer_ShouldConsiderTierDeadIfAllTrackerFailsWithSingleTracker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var trackers []ITrackerAnnouncer
	t1 := NewMockITrackerAnnouncer(ctrl)
	c1 := make(chan trackerAnnounceResult)
	t1.EXPECT().Responses().Return(c1).AnyTimes()
	t1.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	t1.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t1)

	gomock.InOrder(
		t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).Do(func(anF AnnouncingFunction, e tracker.AnnounceEvent) {
			c1 <- trackerAnnounceResult{Err: errors.New("nop")}
		}).Times(1),
	)

	tier, _ := newFallbackTrackersTierAnnouncer(trackers...)
	go tier.startAnnounceLoop(ThirtyMinutesIntervalNoOpAnnouncingFunc, tracker.Started)
	defer tier.stopAnnounceLoop()

	select {
	case st := <-tier.States():
		if st != DEAD {
			t.Fatalf("should have reported tier DEAD, %v received", st)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("should have reported state before timeout")
	}
}

func Test_FallbackTrackersTierAnnouncer_ShouldReconsiderDeadTierAliveIfOneTrackerSucceed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var trackers []ITrackerAnnouncer
	t1 := NewMockITrackerAnnouncer(ctrl)
	c1 := make(chan trackerAnnounceResult)
	t1.EXPECT().Responses().Return(c1).AnyTimes()
	t1.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	t1.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t1)

	t2 := NewMockITrackerAnnouncer(ctrl)
	c2 := make(chan trackerAnnounceResult)
	t2.EXPECT().Responses().Return(c2).AnyTimes()
	t2.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	t2.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t2)

	DefaultDurationWaitOnError = 1 * time.Millisecond
	gomock.InOrder(
		t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).Do(func(anF AnnouncingFunction, e tracker.AnnounceEvent) {
			c1 <- trackerAnnounceResult{Err: errors.New("nop")}
		}).Times(1),
		t2.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).Do(func(anF AnnouncingFunction, e tracker.AnnounceEvent) {
			c2 <- trackerAnnounceResult{Err: errors.New("nop")}
		}).Times(1),
		t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).Do(func(anF AnnouncingFunction, e tracker.AnnounceEvent) {
			c1 <- trackerAnnounceResult{Interval: 1800 * time.Second, Completed: time.Now()}
		}).Times(1),
	)

	tier, _ := newFallbackTrackersTierAnnouncer(trackers...)
	go tier.startAnnounceLoop(ThirtyMinutesIntervalNoOpAnnouncingFunc, tracker.Started)
	defer tier.stopAnnounceLoop()

	select {
	case st := <-tier.States():
		if st != DEAD {
			t.Fatalf("should have reported tier DEAD first, %v received", st)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("should have reported state before timeout")
	}

	select {
	case st := <-tier.States():
		if st != ALIVE {
			t.Fatalf("should have reported tier ALIVE, %v received", st)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("should have reported state before timeout")
	}
}

func Test_FallbackTrackersTierAnnouncer_ShouldNotPreventStopIfATrackerIsTakingForeverToAnnounce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var trackers []ITrackerAnnouncer
	t1 := NewMockITrackerAnnouncer(ctrl)
	c1 := make(chan trackerAnnounceResult)
	t1.EXPECT().Responses().Return(c1).AnyTimes()
	t1.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	t1.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t1)

	latch := congo.NewCountDownLatch(1)
	gomock.InOrder(
		t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).Do(func(anF AnnouncingFunction, e tracker.AnnounceEvent) {
			latch.CountDown()
			time.Sleep(50 * time.Hour)
		}).Times(1),
	)

	tier, _ := newFallbackTrackersTierAnnouncer(trackers...)
	go tier.startAnnounceLoop(ThirtyMinutesIntervalNoOpAnnouncingFunc, tracker.Started)

	if !latch.WaitTimeout(50 * time.Millisecond) {
		t.Fatal("latch has not released")
	}
	tier.stopAnnounceLoop()
}

func Test_FallbackTrackersTierAnnouncer_ShouldBeSafeToRunWithTremendousAmountOfTrackers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	latch := congo.NewCountDownLatch(10000)
	var trackers []ITrackerAnnouncer

	DefaultDurationWaitOnError = 0 * time.Millisecond
	for i := 0; i < 3000; i++ {
		t := NewMockITrackerAnnouncer(ctrl)
		c := make(chan trackerAnnounceResult)
		t.EXPECT().Responses().Return(c).AnyTimes()
		t.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
		t.EXPECT().stopAnnounceLoop().AnyTimes()
		t.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).Do(func(anF AnnouncingFunction, e tracker.AnnounceEvent) {
			defer latch.CountDown()
			c <- trackerAnnounceResult{Err: errors.New("nop")}
		}).MinTimes(1)

		trackers = append(trackers, t)
	}

	tier, _ := newFallbackTrackersTierAnnouncer(trackers...)
	go tier.startAnnounceLoop(ThirtyMinutesIntervalNoOpAnnouncingFunc, tracker.Started)

	if !latch.WaitTimeout(5000 * time.Millisecond) {
		t.Fatal("latch has not released")
	}
	tier.stopAnnounceLoop()
}

func Test_FallbackTrackersTierAnnouncer_ShouldFailToBuildWithEmptyTrackerList(t *testing.T) {
	_, err := newFallbackTrackersTierAnnouncer()
	if err == nil {
		t.Fatal("Should have failed to build")
	}
}

func Test_FallbackTrackersTierAnnouncer_ShouldNotBlockWhenStopAnnounceLoopIsCalledButTheTierWasNotStarted(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tier, _ := newFallbackTrackersTierAnnouncer(NewMockITrackerAnnouncer(ctrl))

	c := make(chan struct{})
	go func() {
		tier.stopAnnounceLoop()
		tier.stopAnnounceLoop()
		tier.stopAnnounceLoop()
		tier.stopAnnounceLoop()
		close(c)
	}()

	select {
	case <-c:
	case <-time.After(50 * time.Millisecond):
		t.Fatal("timeout reached")
	}
}

func Test_FallbackTrackersTierAnnouncer_ShouldCallTrackerOneByOneTillOneSucceed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var trackers []ITrackerAnnouncer
	t1 := NewMockITrackerAnnouncer(ctrl)
	c1 := make(chan trackerAnnounceResult)
	t1.EXPECT().Responses().Return(c1).AnyTimes()
	t1.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	t1.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t1)

	t2 := NewMockITrackerAnnouncer(ctrl)
	c2 := make(chan trackerAnnounceResult)
	t2.EXPECT().Responses().Return(c2).AnyTimes()
	t2.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	t2.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t2)

	latch := congo.NewCountDownLatch(1)
	gomock.InOrder(
		t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).Do(func(anF AnnouncingFunction, e tracker.AnnounceEvent) {
			c1 <- trackerAnnounceResult{Err: errors.New("nop")}
		}).Times(1),
		t2.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).Do(func(anF AnnouncingFunction, e tracker.AnnounceEvent) {
			defer latch.CountDown()
			c2 <- trackerAnnounceResult{Interval: 1800 * time.Second, Completed: time.Now()}
		}).Times(1),
	)

	tier, _ := newFallbackTrackersTierAnnouncer(trackers...)
	go tier.startAnnounceLoop(ThirtyMinutesIntervalNoOpAnnouncingFunc, tracker.Started)

	if !latch.WaitTimeout(50 * time.Millisecond) {
		t.Fatal("latch has not released")
	}
	tier.stopAnnounceLoop()
}

func Test_FallbackTrackersTierAnnouncer_ShouldCallTrackerOneByOneTillOneSucceedUpToLast(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var trackers []ITrackerAnnouncer
	t1 := NewMockITrackerAnnouncer(ctrl)
	c1 := make(chan trackerAnnounceResult)
	t1.EXPECT().Responses().Return(c1).AnyTimes()
	t1.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	t1.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t1)

	t2 := NewMockITrackerAnnouncer(ctrl)
	c2 := make(chan trackerAnnounceResult)
	t2.EXPECT().Responses().Return(c2).AnyTimes()
	t2.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	t2.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t2)

	t3 := NewMockITrackerAnnouncer(ctrl)
	c3 := make(chan trackerAnnounceResult)
	t3.EXPECT().Responses().Return(c3).AnyTimes()
	t3.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	t3.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t3)

	latch := congo.NewCountDownLatch(1)
	gomock.InOrder(
		t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).Do(func(anF AnnouncingFunction, e tracker.AnnounceEvent) {
			c1 <- trackerAnnounceResult{Err: errors.New("nop")}
		}).Times(1),
		t2.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).Do(func(anF AnnouncingFunction, e tracker.AnnounceEvent) {
			c2 <- trackerAnnounceResult{Err: errors.New("nop")}
		}).Times(1),
		t3.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).Do(func(anF AnnouncingFunction, e tracker.AnnounceEvent) {
			defer latch.CountDown()
			c3 <- trackerAnnounceResult{Interval: 1800 * time.Second, Completed: time.Now()}
		}).Times(1),
	)

	tier, _ := newFallbackTrackersTierAnnouncer(trackers...)
	go tier.startAnnounceLoop(ThirtyMinutesIntervalNoOpAnnouncingFunc, tracker.Started)

	if !latch.WaitTimeout(50 * time.Millisecond) {
		t.Fatal("latch has not released")
	}
	tier.stopAnnounceLoop()
}

func Test_FallbackTrackersTierAnnouncer_ShouldStopTrackerBeforeMovingToNext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var trackers []ITrackerAnnouncer
	t1 := NewMockITrackerAnnouncer(ctrl)
	c1 := make(chan trackerAnnounceResult)
	t1.EXPECT().Responses().Return(c1).AnyTimes()
	t1.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	trackers = append(trackers, t1)

	t2 := NewMockITrackerAnnouncer(ctrl)
	c2 := make(chan trackerAnnounceResult)
	t2.EXPECT().Responses().Return(c2).AnyTimes()
	t2.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	t2.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t2)

	latch := congo.NewCountDownLatch(1)
	gomock.InOrder(
		t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).Do(func(anF AnnouncingFunction, e tracker.AnnounceEvent) {
			c1 <- trackerAnnounceResult{Err: errors.New("nop")}
		}).Times(1),
		t1.EXPECT().stopAnnounceLoop().Times(1),
		t2.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).Do(func(anF AnnouncingFunction, e tracker.AnnounceEvent) {
			defer latch.CountDown()
			c2 <- trackerAnnounceResult{Interval: 1800 * time.Second, Completed: time.Now()}
		}).Times(1),
	)

	tier, _ := newFallbackTrackersTierAnnouncer(trackers...)
	go tier.startAnnounceLoop(ThirtyMinutesIntervalNoOpAnnouncingFunc, tracker.Started)

	if !latch.WaitTimeout(50 * time.Millisecond) {
		t.Fatal("latch has not released")
	}
	tier.stopAnnounceLoop()
}

func Test_FallbackTrackersTierAnnouncer_ShouldReorderTrackerListOnAnnounceSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var trackers []ITrackerAnnouncer
	t1 := NewMockITrackerAnnouncer(ctrl)
	c1 := make(chan trackerAnnounceResult)
	t1.EXPECT().Responses().Return(c1).AnyTimes()
	t1.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	t1.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t1)

	t2 := NewMockITrackerAnnouncer(ctrl)
	c2 := make(chan trackerAnnounceResult)
	t2.EXPECT().Responses().Return(c2).AnyTimes()
	t2.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	t2.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t2)

	t3 := NewMockITrackerAnnouncer(ctrl)
	c3 := make(chan trackerAnnounceResult)
	t3.EXPECT().Responses().Return(c3).AnyTimes()
	t3.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	t3.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t3)

	latch := congo.NewCountDownLatch(1)
	gomock.InOrder(
		t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).Do(func(anF AnnouncingFunction, e tracker.AnnounceEvent) {
			c1 <- trackerAnnounceResult{Err: errors.New("nop")}
		}).Times(1),
		t2.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).Do(func(anF AnnouncingFunction, e tracker.AnnounceEvent) {
			c2 <- trackerAnnounceResult{Interval: 1 * time.Millisecond, Completed: time.Now()}
			c2 <- trackerAnnounceResult{Err: errors.New("nop")} // send a success then send an error
		}).Times(1),
		// t2 has succeed, the order should now be t2, t1, t3. When t2 fails it should go to t1
		t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).Do(func(anF AnnouncingFunction, e tracker.AnnounceEvent) {
			defer latch.CountDown()
			c1 <- trackerAnnounceResult{Interval: 1800 * time.Second, Completed: time.Now()}
		}).Times(1),
	)

	tier, _ := newFallbackTrackersTierAnnouncer(trackers...)
	go tier.startAnnounceLoop(ThirtyMinutesIntervalNoOpAnnouncingFunc, tracker.Started)

	if !latch.WaitTimeout(50 * time.Millisecond) {
		t.Fatal("latch has not released")
	}
	tier.stopAnnounceLoop()
}

func Test_FallbackTrackersTierAnnouncer_ShouldNotReportAliveAfterFirstAnnounceFailedButOtherNotAnswered(t *testing.T) {
	var trackers []ITrackerAnnouncer

	for i := 0; i < 30; i++ {
		trackerUrl := testutils.MustParseUrl(fmt.Sprintf("http://localhost/%d", i))
		trackers = append(trackers, newTracker(*trackerUrl))
	}

	tier, _ := newFallbackTrackersTierAnnouncer(trackers...)
	lock := &sync.Mutex{}
	latch := congo.NewCountDownLatch(1)
	var annFunc = buildErrAnnouncingFunc(func(u url.URL) {
		lock.Lock() // first call lock the mutex to prevent any other announce to run
		defer latch.CountDown()
	})

	go tier.startAnnounceLoop(annFunc, tracker.Started)
	defer tier.stopAnnounceLoop()

	if !latch.WaitTimeout(50 * time.Millisecond) {
		t.Fatalf("should have released the latch")
	}

	select {
	case <-time.After(50 * time.Millisecond):
		// perfect, it has not reported his state
	case <-tier.States():
		t.Fatalf("should not have reported his state yet")
	}
}

func Test_FallbackTrackersTierAnnouncer_ShouldReportAliveAfterFirstAnnounceSuccess(t *testing.T) {
	var trackers []ITrackerAnnouncer

	for i := 0; i < 30; i++ {
		trackerUrl := testutils.MustParseUrl(fmt.Sprintf("http://localhost/%d", i))
		trackers = append(trackers, newTracker(*trackerUrl))
	}

	tier, _ := newFallbackTrackersTierAnnouncer(trackers...)
	lock := &sync.Mutex{}
	latch := congo.NewCountDownLatch(1)
	var annFunc = buildAnnouncingFunc(1800*time.Second, func(u url.URL) {
		lock.Lock() // first call lock the mutex to prevent any other announce to run
		defer latch.CountDown()
	})

	go tier.startAnnounceLoop(annFunc, tracker.Started)
	defer tier.stopAnnounceLoop()

	if !latch.WaitTimeout(50 * time.Millisecond) {
		t.Fatalf("should have released the latch")
	}

	select {
	case <-time.After(50 * time.Millisecond):
		// perfect, it has not reported his state
	case s := <-tier.States():
		if s != ALIVE {
			t.Fatalf("should have reported state ALIVE")
		}
	}
}

func Test_FallbackTrackersTierAnnouncer_ShouldAnnounceOnceToFirstTrackerAndReturnIfSucceed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var trackers []ITrackerAnnouncer
	t1 := NewMockITrackerAnnouncer(ctrl)
	c1 := make(chan trackerAnnounceResult)
	t1.EXPECT().Responses().Return(c1).AnyTimes()
	t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).AnyTimes()
	t1.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t1)

	t2 := NewMockITrackerAnnouncer(ctrl)
	c2 := make(chan trackerAnnounceResult)
	t2.EXPECT().Responses().Return(c2).AnyTimes()
	t2.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).AnyTimes()
	t2.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t2)

	t3 := NewMockITrackerAnnouncer(ctrl)
	c3 := make(chan trackerAnnounceResult)
	t3.EXPECT().Responses().Return(c3).AnyTimes()
	t3.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).AnyTimes()
	t3.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t3)

	tier, _ := newFallbackTrackersTierAnnouncer(trackers...)

	t1.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).Return(trackerAnnounceResult{Err: nil}).Times(1)
	t2.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	t3.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	state := tier.announceOnce(context.Background(), nil, tracker.Started)
	if state != ALIVE {
		t.Fatal("should have returned tier alive")
	}
}

func Test_FallbackTrackersTierAnnouncer_ShouldAnnounceOnceByFallingBackUntilOnceSucceed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var trackers []ITrackerAnnouncer
	t1 := NewMockITrackerAnnouncer(ctrl)
	c1 := make(chan trackerAnnounceResult)
	t1.EXPECT().Responses().Return(c1).AnyTimes()
	t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).AnyTimes()
	t1.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t1)

	t2 := NewMockITrackerAnnouncer(ctrl)
	c2 := make(chan trackerAnnounceResult)
	t2.EXPECT().Responses().Return(c2).AnyTimes()
	t2.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).AnyTimes()
	t2.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t2)

	t3 := NewMockITrackerAnnouncer(ctrl)
	c3 := make(chan trackerAnnounceResult)
	t3.EXPECT().Responses().Return(c3).AnyTimes()
	t3.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).AnyTimes()
	t3.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t3)

	tier, _ := newFallbackTrackersTierAnnouncer(trackers...)

	t1.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).Return(trackerAnnounceResult{Err: errors.New("nop")}).Times(1)
	t2.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).Return(trackerAnnounceResult{Err: errors.New("nop")}).Times(1)
	t3.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).Return(trackerAnnounceResult{Err: nil}).Times(1)

	state := tier.announceOnce(context.Background(), nil, tracker.Started)
	if state != ALIVE {
		t.Fatal("should have returned tier alive")
	}
}

func Test_FallbackTrackersTierAnnouncer_ShouldAnnounceOnceByFallingBackAndReturnDeadIfNoneSucceed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var trackers []ITrackerAnnouncer
	t1 := NewMockITrackerAnnouncer(ctrl)
	c1 := make(chan trackerAnnounceResult)
	t1.EXPECT().Responses().Return(c1).AnyTimes()
	t1.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).AnyTimes()
	t1.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t1)

	t2 := NewMockITrackerAnnouncer(ctrl)
	c2 := make(chan trackerAnnounceResult)
	t2.EXPECT().Responses().Return(c2).AnyTimes()
	t2.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).AnyTimes()
	t2.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t2)

	t3 := NewMockITrackerAnnouncer(ctrl)
	c3 := make(chan trackerAnnounceResult)
	t3.EXPECT().Responses().Return(c3).AnyTimes()
	t3.EXPECT().startAnnounceLoop(gomock.Any(), gomock.Any()).AnyTimes()
	t3.EXPECT().stopAnnounceLoop().AnyTimes()
	trackers = append(trackers, t3)

	tier, _ := newFallbackTrackersTierAnnouncer(trackers...)

	t1.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).Return(trackerAnnounceResult{Err: errors.New("nop")}).Times(1)
	t2.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).Return(trackerAnnounceResult{Err: errors.New("nop")}).Times(1)
	t3.EXPECT().announceOnce(gomock.Any(), gomock.Any(), gomock.Any()).Return(trackerAnnounceResult{Err: errors.New("nop")}).Times(1)

	state := tier.announceOnce(context.Background(), nil, tracker.Started)
	if state != DEAD {
		t.Fatal("should have returned tier dead")
	}
}
