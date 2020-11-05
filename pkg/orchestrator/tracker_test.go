package orchestrator

import (
	"context"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/announcer"
	"github.com/anthonyraymond/joal-cli/pkg/utils/testutils"
	"github.com/golang/mock/gomock"
	"github.com/nvn1729/congo"
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
	"time"
)

type mockedTrackerAnnouncer struct {
	annOnce      func(ctx context.Context, announce AnnouncingFunction, event tracker.AnnounceEvent) trackerAnnounceResult
	startAnnLoop func(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) trackerAnnounceResult
	stopAnnLoop  func()
	c            chan trackerAnnounceResult
}

func (m *mockedTrackerAnnouncer) announceOnce(ctx context.Context, announce AnnouncingFunction, event tracker.AnnounceEvent) trackerAnnounceResult {
	if m.annOnce != nil {
		return m.annOnce(ctx, announce, event)
	}
	return trackerAnnounceResult{}
}

func (m *mockedTrackerAnnouncer) startAnnounceLoop(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) {
	if m.startAnnLoop != nil {
		m.startAnnLoop(announce, firstEvent)
	}
}

func (m *mockedTrackerAnnouncer) Responses() <-chan trackerAnnounceResult {
	return m.c
}

func (m *mockedTrackerAnnouncer) stopAnnounceLoop() {
	if m.stopAnnLoop != nil {
		m.stopAnnLoop()
	}
}

func Test_TrackerAnnouncer_ShouldChangeNextAnnounceToNoneIfFirsAnnounceIsStarted(t *testing.T) {
	var announceEvents []tracker.AnnounceEvent
	latch := congo.NewCountDownLatch(2)
	//noinspection GoVarAndConstTypeMayBeOmitted
	var annFunc AnnouncingFunction = func(ctx context.Context, u url.URL, event tracker.AnnounceEvent) (announcer.AnnounceResponse, error) {
		defer func() { _ = latch.CountDown() }()
		announceEvents = append(announceEvents, event)
		return announcer.AnnounceResponse{
			Interval: 1 * time.Millisecond,
		}, nil
	}

	tra := newTracker(url.URL{})

	go tra.startAnnounceLoop(annFunc, tracker.Started)
	if !latch.WaitTimeout(2 * time.Second) {
		t.Fatal("Latch has not been released")
	}
	tra.stopAnnounceLoop()

	assert.Equal(t, announceEvents[0], tracker.Started)
	assert.Equal(t, announceEvents[1], tracker.None)
}

func Test_TrackerAnnouncer_AnnounceStartLoopShouldReturnAfterStop(t *testing.T) {
	var announceEvents []tracker.AnnounceEvent
	announceLatch := congo.NewCountDownLatch(1)
	endedLatch := congo.NewCountDownLatch(1)
	//noinspection GoVarAndConstTypeMayBeOmitted
	var annFunc AnnouncingFunction = func(ctx context.Context, u url.URL, event tracker.AnnounceEvent) (announcer.AnnounceResponse, error) {
		defer func() { _ = announceLatch.CountDown() }()
		announceEvents = append(announceEvents, event)
		return announcer.AnnounceResponse{
			Interval: 1 * time.Millisecond,
		}, nil
	}

	tra := newTracker(url.URL{})

	go func() {
		defer endedLatch.CountDown()
		tra.startAnnounceLoop(annFunc, tracker.None)
	}()
	if !announceLatch.WaitTimeout(1 * time.Second) {
		t.Fatal("Latch has not been released")
	}
	tra.stopAnnounceLoop()
	// if this does not release, the defer has not been called, and it means that the startAnnounceLoop has not returned after stop
	if !endedLatch.WaitTimeout(1 * time.Second) {
		t.Fatal("Latch has not been released")
	}

	assert.Equal(t, announceEvents[0], tracker.None)
}

func Test_TrackerAnnouncer_ShouldBeReusableAfterStopLoop(t *testing.T) {
	var announceEvents []tracker.AnnounceEvent
	announceLatch := congo.NewCountDownLatch(1)
	//noinspection GoVarAndConstTypeMayBeOmitted
	var annFunc AnnouncingFunction = func(ctx context.Context, u url.URL, event tracker.AnnounceEvent) (announcer.AnnounceResponse, error) {
		defer func() { _ = announceLatch.CountDown() }()
		announceEvents = append(announceEvents, event)
		return announcer.AnnounceResponse{
			Interval: 1 * time.Millisecond,
		}, nil
	}

	tra := newTracker(url.URL{})

	go tra.startAnnounceLoop(annFunc, tracker.None)
	if !announceLatch.WaitTimeout(1 * time.Second) {
		t.Fatal("Latch has not been released")
	}
	tra.stopAnnounceLoop()

	numberOfAnnounceFirstTime := len(announceEvents)
	*announceLatch = *congo.NewCountDownLatch(1)

	go tra.startAnnounceLoop(annFunc, tracker.None)
	if !announceLatch.WaitTimeout(1 * time.Second) {
		t.Fatal("Latch has not been released")
	}
	tra.stopAnnounceLoop()

	assert.Greater(t, len(announceEvents), numberOfAnnounceFirstTime)
}

func Test_TrackerAnnouncer_ShouldFeedChannelWithResponse(t *testing.T) {
	tra := newTracker(url.URL{})

	go tra.startAnnounceLoop(ZeroIntervalNoOpAnnouncingFunc, tracker.None)
	defer tra.stopAnnounceLoop()

	done := make(chan struct{})
	go func() {
		<-tra.Responses()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout")
	}
}

func Test_TrackerAnnouncer_ShouldNotBlockWhenStopAnnounceLoopIsCalledButTheTrackerWasNotStarted(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tr := newTracker(*testutils.MustParseUrl("http://localhost"))

	latch := congo.NewCountDownLatch(1)
	go func() {
		tr.stopAnnounceLoop()
		tr.stopAnnounceLoop()
		tr.stopAnnounceLoop()
		tr.stopAnnounceLoop()
		tr.stopAnnounceLoop()
		tr.stopAnnounceLoop()
		latch.CountDown()
	}()

	if !latch.WaitTimeout(500 * time.Millisecond) {
		t.Fatal("Should not have blocked")
	}
}

func Test_TrackerAnnouncer_ShouldAnnounceOnce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tr := newTracker(*testutils.MustParseUrl("http://localhost"))

	latch := congo.NewCountDownLatch(1)

	var annFunc = buildAnnouncingFunc(1800*time.Second, func(u url.URL) {
		latch.CountDown()
	})
	go tr.announceOnce(context.Background(), annFunc, tracker.Started)

	if !latch.WaitTimeout(50 * time.Millisecond) {
		t.Fatal("timed out")
	}
}
