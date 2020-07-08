package orchestrator

import (
	"context"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/internal/testutils"
	"github.com/golang/mock/gomock"
	"github.com/nvn1729/congo"
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
	"time"
)

func Test_TrackerAnnouncer_ShouldChangeNextAnnounceToNoneIfFirsAnnounceIsStarted(t *testing.T) {
	var announceEvents []tracker.AnnounceEvent
	latch := congo.NewCountDownLatch(2)
	var annFunc AnnouncingFunction = func(u url.URL, event tracker.AnnounceEvent, ctx context.Context) (tracker.AnnounceResponse, error) {
		defer func() { _ = latch.CountDown() }()
		announceEvents = append(announceEvents, event)
		return tracker.AnnounceResponse{
			Interval: 1,
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
	var annFunc AnnouncingFunction = func(u url.URL, event tracker.AnnounceEvent, ctx context.Context) (tracker.AnnounceResponse, error) {
		defer func() { _ = announceLatch.CountDown() }()
		announceEvents = append(announceEvents, event)
		return tracker.AnnounceResponse{
			Interval: 1,
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
	var annFunc AnnouncingFunction = func(u url.URL, event tracker.AnnounceEvent, ctx context.Context) (tracker.AnnounceResponse, error) {
		defer func() { _ = announceLatch.CountDown() }()
		announceEvents = append(announceEvents, event)
		return tracker.AnnounceResponse{
			Interval: 1,
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
	var annFunc AnnouncingFunction = func(u url.URL, event tracker.AnnounceEvent, ctx context.Context) (tracker.AnnounceResponse, error) {
		return tracker.AnnounceResponse{Interval: 1}, nil
	}

	tra := newTracker(url.URL{})

	go tra.startAnnounceLoop(annFunc, tracker.None)
	defer tra.stopAnnounceLoop()

	done := make(chan struct{})
	go func() {
		<-tra.Responses()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(50 * time.Millisecond):
		t.Fatal("timedout")
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

	var annFunc AnnouncingFunction = func(u url.URL, event tracker.AnnounceEvent, ctx context.Context) (tracker.AnnounceResponse, error) {
		defer latch.CountDown()
		return tracker.AnnounceResponse{}, nil
	}
	go tr.announceOnce(annFunc, tracker.Started)

	if !latch.WaitTimeout(50 * time.Millisecond) {
		t.Fatal("timed out")
	}
}