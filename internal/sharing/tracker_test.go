package sharing

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"math"
	"net/url"
	"testing"
	"time"
)

func mustParseUrl(str string) *url.URL {
	u, err := url.Parse(str)
	if err != nil {
		panic(err)
	}
	return u
}

func defaultTracker() Tracker {
	return newTracker(mustParseUrl("http://localhost:8080"))
}

func TestTracker_shouldReceiveAnnounceResponse(t *testing.T) {
	tracker := defaultTracker()
	announce, _ := tracker.announce(Started)

	err := tracker.announceSucceed(TrackerAnnounceResponse{
		announceUid: announce.uid,
		Interval:    1800 * time.Second,
	})

	assert.NoError(t, err)
	assert.Equal(t, 0, tracker.ConsecutiveFails())                             // reset consecutive fails ont success
	assert.False(t, tracker.requireAnnounce(time.Now().Add(1790*time.Second))) // should not be able to announce before 1800s
	assert.True(t, tracker.requireAnnounce(time.Now().Add(1801*time.Second)))
}

func TestTracker_shouldDenyAnnounceIfNotPending(t *testing.T) {
	tracker := defaultTracker()

	err := tracker.announceSucceed(TrackerAnnounceResponse{announceUid: uuid.New()})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected announce response")
}

func TestTracker_shouldReceiveAnnounceError(t *testing.T) {
	tracker := defaultTracker()
	announce, _ := tracker.announce(Started)

	err := tracker.announceFailed(TrackerAnnounceError{announceUid: announce.uid})

	assert.NoError(t, err)
	assert.Equal(t, 1, tracker.ConsecutiveFails())
}

func TestTracker_shouldDenyAnnounceErrorIfNotPending(t *testing.T) {
	tracker := defaultTracker()

	err := tracker.announceFailed(TrackerAnnounceError{announceUid: uuid.New()})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected announce response")
}

func TestTracker_shouldDisableTracker(t *testing.T) {
	tracker := defaultTracker()

	assert.False(t, tracker.IsDisabled())
	tracker.disable(AnnounceProtocolNotSupported)
	assert.True(t, tracker.IsDisabled())
}

func TestTracker_shouldDelayNextAnnounceMoreAndMoreAsConsecutiveFailsIncrease(t *testing.T) {
	tracker := defaultTracker()

	var announcesDelay []time.Duration
	for i := 0; i < 5; i++ {
		req, _ := tracker.announce(None)
		_ = tracker.announceFailed(TrackerAnnounceError{announceUid: req.uid})
		announcesDelay = append(announcesDelay, tracker.NextAnnounceAt().Sub(time.Now()))
	}

	for i := 0; i < 4; i++ {
		assert.Less(t, announcesDelay[i], announcesDelay[i+1])
	}
}

func TestTracker_requireAnnounce(t *testing.T) {
	assert.True(t, (&trackerImpl{}).requireAnnounce(time.Now()))
	assert.True(t, (&trackerImpl{nextAnnounceAt: time.Now().Add(5 * time.Minute)}).requireAnnounce(time.Now().Add(6*time.Minute)))

	assert.False(t, (&trackerImpl{pendingAnnounce: &trackerAnnounceRequest{}}).requireAnnounce(time.Now()), "can not announce when already announcing")
	assert.False(t, (&trackerImpl{nextAnnounceAt: time.Now().Add(5 * time.Hour)}).requireAnnounce(time.Now()), "can not announce when nextAnnounce is after")
	assert.False(t, (&trackerImpl{nextAnnounceAt: time.Now().Add(-5 * time.Hour)}).requireAnnounce(time.Now().Add(-6*time.Hour)), "can not announce when nextAnnounce is after")
	assert.False(t, (&trackerImpl{disabled: TrackerDisabled{isDisabled: true}}).requireAnnounce(time.Now()), "can not announce when disabled")
}

func TestTracker_shouldAnnounce(t *testing.T) {
	tracker := defaultTracker()

	req, err := tracker.announce(Started)

	assert.NoError(t, err)
	assert.Len(t, req.uid.String(), 36)
	assert.Equal(t, tracker.Url(), req.url)
	assert.Equal(t, Started, req.event)
}

func TestTracker_announceShouldReplaceNoneWithStatedIfNeverAnnouncedStarted(t *testing.T) {
	tracker := defaultTracker()

	req, err := tracker.announce(None)

	assert.NoError(t, err)
	assert.Equal(t, Started, req.event)
}

func TestTracker_shouldPreventStartedAnnounceWhenAwaitingAnotherAnswer(t *testing.T) {
	tracker := defaultTracker()

	_, _ = tracker.announce(Started)
	_, err := tracker.announce(Started)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "awaiting for an announce response")
}

func TestTracker_shouldPreventNoneAnnounceWhenAwaitingAnotherAnswer(t *testing.T) {
	tracker := defaultTracker()

	_, _ = tracker.announce(Started)
	_, err := tracker.announce(None)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "awaiting for an announce response")
}

func TestTracker_shouldAllowCompletedAnnounceWhenAwaitingAnotherAnswer(t *testing.T) {
	tracker := defaultTracker()

	startedAnnounce, _ := tracker.announce(Started)
	completedAnnounce, err := tracker.announce(Completed)
	assert.NoError(t, err, "should allow Completed announce even if another is queued")

	err = tracker.announceSucceed(TrackerAnnounceResponse{announceUid: startedAnnounce.uid})
	assert.Error(t, err, "started announce should have been replaced and answer should be ignored")
	assert.Contains(t, err.Error(), "unexpected announce response")

	err = tracker.announceSucceed(TrackerAnnounceResponse{announceUid: completedAnnounce.uid})
	assert.NoError(t, err)
}

func TestTracker_shouldAllowStoppedAnnounceWhenAwaitingAnotherAnswer(t *testing.T) {
	tracker := defaultTracker()

	startedAnnounce, _ := tracker.announce(Started)
	stoppedAnnounce, err := tracker.announce(Stopped)
	assert.NoError(t, err, "should allow Stopped announce even if another is queued")

	err = tracker.announceSucceed(TrackerAnnounceResponse{announceUid: startedAnnounce.uid})
	assert.Error(t, err, "started announce should have been replaced and answer should be ignored")
	assert.Contains(t, err.Error(), "unexpected announce response")

	err = tracker.announceSucceed(TrackerAnnounceResponse{announceUid: stoppedAnnounce.uid})
	assert.NoError(t, err)
}

func TestCalculateBackoff(t *testing.T) {
	minDuration := 5 * time.Second
	maxDuration := 1800 * time.Second

	var values []time.Duration
	for i := 0; i < 16; i++ {
		values = append(values, calculateBackoff(i, minDuration, maxDuration))
	}

	expected := []time.Duration{
		minDuration,
		15 * time.Second,
		45 * time.Second,
		95 * time.Second,
		165 * time.Second,
		255 * time.Second,
		365 * time.Second,
		495 * time.Second,
		645 * time.Second,
		815 * time.Second,
		1005 * time.Second,
		1215 * time.Second,
		1445 * time.Second,
		1695 * time.Second,
		maxDuration,
		maxDuration,
	}

	assert.Equal(t, expected, values)
}

func TestCalculateBackoff_shouldNotReturnOverflowedValue(t *testing.T) {
	minDuration := 5 * time.Second
	maxDuration := 1800 * time.Second

	res := calculateBackoff(math.MaxInt, minDuration, maxDuration)

	assert.GreaterOrEqual(t, res, minDuration)
	assert.LessOrEqual(t, res, maxDuration)
}
