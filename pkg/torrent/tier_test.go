package torrent

import (
	"context"
	"errors"
	"fmt"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/internal/testutils"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/nvn1729/congo"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

func buildAnnouncingFunc(interval time.Duration, callbacks ...func(u url.URL)) AnnouncingFunction {
	return func(u url.URL, event tracker.AnnounceEvent, ctx context.Context) trackerAnnounceResult {
		for _, c := range callbacks {
			c(u)
		}
		return trackerAnnounceResult{Err: nil, Interval: interval, Completed: time.Now()}
	}
}

var OneMinuteIntervalAnnouncingFUnc = buildAnnouncingFunc(1 * time.Minute)

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

		t.EXPECT().Uuid().Return(uuid.New()).AnyTimes()
		t.EXPECT().Responses().Return(make(chan trackerAwareAnnounceResult)).AnyTimes()

		trackers = append(trackers, t)
	}

	tier, _ := newAllTrackersTierAnnouncer(trackers...)
	go tier.startAnnounceLoop(OneMinuteIntervalAnnouncingFUnc, tracker.Started)
	defer tier.stopAnnounceLoop()

	if testutils.WaitOrFailAfterTimeout(&wg, 50*time.Millisecond) != nil {
		t.Fatal("not ALL the trackers have been instruct to announce")
	}
}

func Test_AllTrackersTierAnnouncer_ShouldBeReusableAfterStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var trackers []ITrackerAnnouncer
	var wg sync.WaitGroup

	for i := 0; i < 30; i++ {
		trackerUrl := testutils.MustParseUrl(fmt.Sprintf("http://localhost/%d", i))
		trackers = append(trackers, newTracker(*trackerUrl))
	}

	tier, _ := newAllTrackersTierAnnouncer(trackers...)
	var annFunc AnnouncingFunction = buildAnnouncingFunc(1*time.Minute, func(u url.URL) { wg.Done() })

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var trackers []ITrackerAnnouncer

	for i := 0; i < 30; i++ {
		trackerUrl := testutils.MustParseUrl(fmt.Sprintf("http://localhost/%d", i))
		trackers = append(trackers, newTracker(*trackerUrl))
	}

	tier, _ := newAllTrackersTierAnnouncer(trackers...)
	var failAnnFunc AnnouncingFunction = func(u url.URL, event tracker.AnnounceEvent, ctx context.Context) trackerAnnounceResult {
		return trackerAnnounceResult{Err: errors.New("failed")}
	}

	go tier.startAnnounceLoop(failAnnFunc, tracker.Started)
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

func Test_AllTrackersTierAnnouncer_ShouldConsiderTierDeadIfAllTrackerFailsWithSingleTracker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var trackers []ITrackerAnnouncer

	for i := 0; i < 1; i++ {
		trackerUrl := testutils.MustParseUrl(fmt.Sprintf("http://localhost/%d", i))
		trackers = append(trackers, newTracker(*trackerUrl))
	}

	tier, _ := newAllTrackersTierAnnouncer(trackers...)
	var failAnnFunc AnnouncingFunction = func(u url.URL, event tracker.AnnounceEvent, ctx context.Context) trackerAnnounceResult {
		return trackerAnnounceResult{Err: errors.New("failed")}
	}

	go tier.startAnnounceLoop(failAnnFunc, tracker.Started)
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var trackers []ITrackerAnnouncer

	for i := 0; i < 10; i++ {
		trackerUrl := testutils.MustParseUrl(fmt.Sprintf("http://localhost/%d", i))
		trackers = append(trackers, newTracker(*trackerUrl))
	}

	DefaultDurationWaitOnError = 1 * time.Millisecond
	announceResponse := &trackerAnnounceResult{Err: errors.New("failed")}
	tier, _ := newAllTrackersTierAnnouncer(trackers...)
	var failAnnFunc AnnouncingFunction = func(u url.URL, event tracker.AnnounceEvent, ctx context.Context) trackerAnnounceResult {
		fmt.Println("response")
		return *announceResponse
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

	announceResponse = &trackerAnnounceResult{Interval: 1800 * time.Second, Completed: time.Now(), Err: nil}

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var trackers []ITrackerAnnouncer

	for i := 0; i < 10; i++ {
		trackerUrl := testutils.MustParseUrl(fmt.Sprintf("http://localhost/%d", i))
		trackers = append(trackers, newTracker(*trackerUrl))
	}
	latch := congo.NewCountDownLatch(uint(len(trackers)))

	tier, _ := newAllTrackersTierAnnouncer(trackers...)
	var annFunc AnnouncingFunction = buildAnnouncingFunc(1*time.Millisecond, func(u url.URL) {
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var trackers []ITrackerAnnouncer
	var latch *congo.CountDownLatch

	for i := 0; i < 3000; i++ {
		trackerUrl := testutils.MustParseUrl(fmt.Sprintf("http://localhost/%d", i))
		trackers = append(trackers, newTracker(*trackerUrl))
	}

	tier, _ := newAllTrackersTierAnnouncer(trackers...)
	var annFunc AnnouncingFunction = buildAnnouncingFunc(1*time.Millisecond, func(u url.URL) { latch.CountDown() })

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
