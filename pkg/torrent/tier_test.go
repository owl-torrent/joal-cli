package torrent

import (
	"context"
	"fmt"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/internal/testutils"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/nvn1729/congo"
	"net/url"
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

	tier := newAllTrackersTierAnnouncer(trackers...)
	tier.startAnnounceLoop(OneMinuteIntervalAnnouncingFUnc, tracker.Started)
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

	tier := newAllTrackersTierAnnouncer(trackers...)
	var annFunc AnnouncingFunction = buildAnnouncingFunc(1*time.Minute, func(u url.URL) { wg.Done() })

	wg.Add(len(trackers))
	tier.startAnnounceLoop(annFunc, tracker.Started)
	if testutils.WaitOrFailAfterTimeout(&wg, 50*time.Millisecond) != nil {
		t.Fatal("not ALL the trackers have been instruct to announce")
	}
	tier.stopAnnounceLoop()

	wg.Add(len(trackers))
	tier.startAnnounceLoop(annFunc, tracker.Started)
	if testutils.WaitOrFailAfterTimeout(&wg, 50*time.Millisecond) != nil {
		t.Fatal("not ALL the trackers have been instruct to announce")
	}
	tier.stopAnnounceLoop()
}

func Test_AllTrackersTierAnnouncer_ShouldConsiderTierDeadIfAllTrackerFails(t *testing.T) {

}

func Test_AllTrackersTierAnnouncer_ShouldReconsiderDeadTierAliveIfOneTrackerSucceed(t *testing.T) {

}

func Test_AllTrackersTierAnnouncer_ShouldCastTierStateEventOnTrackerAnnounceResponse(t *testing.T) {

}

func Test_AllTrackersTierAnnouncer_ShouldNotPreventStopIfATrackerIsTakingForeverToAnnounce(t *testing.T) {

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

	tier := newAllTrackersTierAnnouncer(trackers...)
	var annFunc AnnouncingFunction = buildAnnouncingFunc(1*time.Millisecond, func(u url.URL) { latch.CountDown() })

	latch = congo.NewCountDownLatch(uint(5 * len(trackers)))
	tier.startAnnounceLoop(annFunc, tracker.Started)
	defer tier.stopAnnounceLoop()

	if !latch.WaitTimeout(500 * time.Millisecond) {
		t.Fatal("not enough announce")
	}
}
