package sharing

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
	"time"
)

/* TODO: tracker impl
 *  - Receive AnnounceResult
 *    - Should calculate nextAnnounceAt (backoff) on announceFailed
 *  - Announce to a tracker
 *    - return an announce request (or announce request builder?)
 *  - canAnnounce()
 *  - Disable a tracker
 *  - Store an history of announce success & error (maybe an object AnnonceHistory{announceResponse, error} where announceResponse and error can be nil ?
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
