package sharing

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"math"
	"net/url"
	"testing"
	"time"
)

//TODO: Tracker
// - if an announce is pending, prevent a None announce to be requested
// - if Completed or Stopped announced is requested, discard any other pending announce

func mustParseUrl(str string) *url.URL {
	u, err := url.Parse(str)
	if err != nil {
		panic(err)
	}
	return u
}

func inTimeSpan(start, end, check time.Time) bool {
	if start.Before(end) {
		return !check.Before(start) && !check.After(end)
	}
	if start.Equal(end) {
		return check.Equal(start)
	}
	return !start.After(check) || !end.Before(check)
}

func TestTracker_shouldReceiveAnnounce(t *testing.T) {
	announceUid := uuid.New()
	tracker := Tracker{
		pendingAnnounce:  &trackerAnnounceRequest{uid: announceUid},
		consecutiveFails: 2,
	}

	tracker.announceSucceed(TrackerAnnounceResponse{
		announceUid: announceUid,
		Interval:    1800 * time.Second,
	})

	assert.Nil(t, tracker.pendingAnnounce)
	assert.Equal(t, 0, tracker.consecutiveFails) // reset consecutive fails ont success
	assert.True(t, inTimeSpan(time.Now().Add(1790*time.Second), time.Now().Add(1810*time.Second), tracker.nextAnnounceAt))
	assert.True(t, tracker.hasAnnouncedOnce)
}

func TestTracker_shouldDenyAnnounceIfNotPending(t *testing.T) {
	tracker := Tracker{
		pendingAnnounce: &trackerAnnounceRequest{uid: uuid.New()},
	}

	tracker.announceSucceed(TrackerAnnounceResponse{announceUid: uuid.New()})

	assert.NotNil(t, tracker.pendingAnnounce)
}

func TestTracker_shouldReceiveAnnounceError(t *testing.T) {
	announceUid := uuid.New()
	tracker := Tracker{
		pendingAnnounce:  &trackerAnnounceRequest{uid: announceUid},
		consecutiveFails: 0,
	}

	tracker.announceFailed(TrackerAnnounceError{announceUid: announceUid})

	assert.Nil(t, tracker.pendingAnnounce)
	assert.Equal(t, 1, tracker.consecutiveFails)
}

func TestTracker_shouldDenyAnnounceErrorIfNotPending(t *testing.T) {
	tracker := Tracker{
		pendingAnnounce: &trackerAnnounceRequest{uid: uuid.New()},
	}

	tracker.announceFailed(TrackerAnnounceError{announceUid: uuid.New()})

	assert.NotNil(t, tracker.pendingAnnounce)
}

func TestTracker_shouldDisableTracker(t *testing.T) {
	tracker := Tracker{}

	assert.False(t, tracker.isDisabled())
	tracker.disable(AnnounceProtocolNotSupported)
	assert.True(t, tracker.isDisabled())
	assert.Equal(t, TrackerDisabled{
		isDisabled: true,
		reason:     AnnounceProtocolNotSupported,
	}, tracker.disabled)
}

func TestTracker_shouldDelayNextAnnounceMoreAndMoreAsConsecutiveFailsIncrease(t *testing.T) {
	tracker := Tracker{
		consecutiveFails: 0,
	}

	var announcesDelay []time.Duration
	for i := 0; i < 5; i++ {
		req := tracker.announce(None)
		tracker.announceFailed(TrackerAnnounceError{announceUid: req.uid})
		announcesDelay = append(announcesDelay, tracker.nextAnnounceAt.Sub(time.Now()))
	}

	for i := 0; i < 4; i++ {
		assert.Less(t, announcesDelay[i], announcesDelay[i+1])
	}
}

func TestTracker_requireAnnounce(t *testing.T) {
	assert.True(t, (&Tracker{}).requireAnnounce(time.Now()))
	assert.True(t, (&Tracker{nextAnnounceAt: time.Now().Add(5 * time.Minute)}).requireAnnounce(time.Now().Add(6*time.Minute)))

	assert.False(t, (&Tracker{pendingAnnounce: &trackerAnnounceRequest{}}).requireAnnounce(time.Now()), "can not announce when already announcing")
	assert.False(t, (&Tracker{nextAnnounceAt: time.Now().Add(5 * time.Hour)}).requireAnnounce(time.Now()), "can not announce when nextAnnounce is after")
	assert.False(t, (&Tracker{nextAnnounceAt: time.Now().Add(-5 * time.Hour)}).requireAnnounce(time.Now().Add(-6*time.Hour)), "can not announce when nextAnnounce is after")
	assert.False(t, (&Tracker{disabled: TrackerDisabled{isDisabled: true}}).requireAnnounce(time.Now()), "can not announce when disabled")
}

func TestTracker_shouldAnnounce(t *testing.T) {
	tracker := Tracker{
		url: mustParseUrl("http://localhost:4333/announce"),
	}

	req := tracker.announce(Started)

	assert.Equal(t, req.uid, tracker.pendingAnnounce.uid)

	assert.Len(t, req.uid.String(), 36)
	assert.Equal(t, tracker.url, req.url)
	assert.Equal(t, Started, req.event)
}

func TestTracker_announceShouldReplaceNoneWithStatedIfNeverAnnouncedStarted(t *testing.T) {
	tracker := Tracker{
		url: mustParseUrl("http://localhost:4333/announce"),
	}

	req := tracker.announce(None)

	assert.Equal(t, Started, req.event)
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
