package sharing

import (
	"github.com/anthonyraymond/joal-cli/pkg/duration"
	"net/url"
	"time"
)

/*
type TrackerState int

const (

	InUse    TrackerState = 0 // InUse describe a tracker that we are announcing to
	Fallback TrackerState = 1 // Fallback describe a tracker that we are not currently announcing to, the tracker is considered a fallback tracker
	Disabled TrackerState = 2 // Disabled describe a disabled tracker that we won't ever use

)
*/

type Tracker struct {
	url              *url.URL
	consecutiveFails int
	isAnnouncing     bool
	nextAnnounceAt   time.Time
	disabled         TrackerDisabled
}

func (t *Tracker) announceSucceed(response TrackerAnnounceResponse) {
	t.isAnnouncing = false
	t.consecutiveFails = 0
	t.nextAnnounceAt = time.Now().Add(response.Interval)
}

func (t *Tracker) announceFailed(error TrackerAnnounceError) {
	t.isAnnouncing = false
	t.nextAnnounceAt = time.Now().Add(calculateBackoff(t.consecutiveFails, 5*time.Second, 1800*time.Second))
	t.consecutiveFails++
}

func (t *Tracker) disable(reason TrackerDisableReason) {
	t.disabled = TrackerDisabled{
		isDisabled: true,
		reason:     reason,
	}
}

func (t *Tracker) isDisabled() bool {
	return t.disabled.isDisabled
}

func (t *Tracker) canAnnounce(at time.Time) bool {
	return !t.isAnnouncing && t.nextAnnounceAt.Before(at) && !t.isDisabled()
}

type TrackerAnnounceResponse struct {
	Interval time.Duration
}

type TrackerAnnounceError struct {
}

type TrackerDisabled struct {
	isDisabled bool
	reason     TrackerDisableReason
}

type TrackerDisableReason struct {
	reason string
}

var (
	AnnounceProtocolNotSupported = TrackerDisableReason{reason: "tracker.disabled.protocol-not-supported"}
)

func calculateBackoff(consecutiveFails int, minDelay time.Duration, maxDelay time.Duration) time.Duration {
	backoffRatio := 250
	// the exponential back-off ends up being:
	// 7, 15, 27, 45, 95, 127, 165, ... seconds
	// with the default tracker_backoff of 250
	sqrt := float64(consecutiveFails * consecutiveFails)

	backoffDelay := minDelay.Seconds() + sqrt*minDelay.Seconds()*float64(backoffRatio/100)

	return duration.Min(
		maxDelay,
		time.Duration(backoffDelay)*time.Second,
	)
}
