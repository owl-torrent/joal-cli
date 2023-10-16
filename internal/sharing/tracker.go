package sharing

import (
	"net/url"
	"time"
)

type TrackerState int

/*
const (

	InUse    TrackerState = 0 // InUse describe a tracker that we are announcing to
	Fallback TrackerState = 1 // Fallback describe a tracker that we are not currently announcing to, the tracker is considered a fallback tracker
	Disabled TrackerState = 2 // Disabled describe a disabled tracker that we won't ever use

)
*/
type Tracker struct {
	state            TrackerState
	url              *url.URL
	consecutiveFails int
	isAnnouncing     bool
	nextAnnounceAt   time.Time
}

func (t *Tracker) announceSucceed(response TrackerAnnounceResponse) {
	t.isAnnouncing = false
	t.consecutiveFails = 0
	t.nextAnnounceAt = time.Now().Add(response.Interval)
}

func (t *Tracker) announceFailed(error TrackerAnnounceError) {
	t.isAnnouncing = false
	t.consecutiveFails++
}

type TrackerAnnounceResponse struct {
	Interval time.Duration
}

type TrackerAnnounceError struct {
}
