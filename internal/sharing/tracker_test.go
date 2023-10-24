package sharing

import (
	"github.com/stretchr/testify/assert"
	"math"
	"net/url"
	"testing"
	"time"
)

/* TODO: tracker impl
 *  - canAnnounce() // not disable, not updating ATM, nextAnnounce > now (now est un time passé en param à la fn)
 *  - Announce to a tracker
 *    - return an announce request (or announce request builder?)
 */

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
	tracker := Tracker{
		isAnnouncing:     true,
		consecutiveFails: 2,
	}

	tracker.announceSucceed(TrackerAnnounceResponse{
		Interval: 1800 * time.Second,
	})

	assert.False(t, tracker.isAnnouncing)
	assert.Equal(t, 0, tracker.consecutiveFails) // reset consecutive fails ont success
	assert.True(t, inTimeSpan(time.Now().Add(1790*time.Second), time.Now().Add(1810*time.Second), tracker.nextAnnounceAt))
}

func TestTracker_shouldReceiveAnnounceError(t *testing.T) {
	tracker := Tracker{
		isAnnouncing:     true,
		consecutiveFails: 0,
	}

	tracker.announceFailed(TrackerAnnounceError{})

	assert.False(t, tracker.isAnnouncing)
	assert.Equal(t, 1, tracker.consecutiveFails)
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
		tracker.announceFailed(TrackerAnnounceError{})
		announcesDelay = append(announcesDelay, tracker.nextAnnounceAt.Sub(time.Now()))
	}

	for i := 0; i < 4; i++ {
		assert.Less(t, announcesDelay[i], announcesDelay[i+1])
	}
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
