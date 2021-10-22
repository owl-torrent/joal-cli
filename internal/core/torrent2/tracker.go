package torrent2

import (
	"math"
	"net/url"
	"time"
)

// Never retry announcing faster than 5s
const trackerRetryDelayMin = 5 * time.Second

// Never wait more than 60min before retrying
const trackerRetryDelayMax = 60 * time.Minute

// Maximum number of announce history to keep for each tracker
const trackerMaxHistorySize = 3

type Tracker interface {
	CanAnnounce(now time.Time) bool
	IsWorking() bool
	Succeed(announceHistory AnnounceHistory)
	Failed(announceHistory AnnounceHistory, backoffRatio int, retryInterval int)
	Reset()
}

type tracker struct {
	url url.URL
	// tier that the tracker belongs to
	tier int16
	// if false the tracker must not be used at all
	enabled      bool
	trackerState *trackerState
}

// CanAnnounce returns true if the tracker is not currently announcing and if the interval to wait is elapsed
func (t *tracker) CanAnnounce(now time.Time) bool {
	// add 1 safety sec before comparing "strictly greater than", because
	if now.Before(t.trackerState.nextAnnounce) {
		return false
	}
	if t.trackerState.updating {
		return false
	}
	return true
}

// IsWorking returns true if the last announce was successful
func (t *tracker) IsWorking() bool {
	return t.trackerState.fails == 0
}

func (t *tracker) Succeed(announceHistory AnnounceHistory) {
	t.trackerState.fails = 0
	t.enqueueAnnounceHistory(announceHistory)

	t.trackerState.nextAnnounce = time.Now().Add(announceHistory.interval)

	t.trackerState.updating = false
}

func (t *tracker) Failed(announceHistory AnnounceHistory, backoffRatio int, retryInterval int) {
	t.trackerState.fails++
	t.enqueueAnnounceHistory(announceHistory)

	// the exponential back-off ends up being:
	// 7, 15, 27, 45, 95, 127, 165, ... seconds
	// with the default tracker_backoff of 250
	failSquare := time.Duration(t.trackerState.fails*t.trackerState.fails) * time.Second

	delay := math.Max(
		float64(retryInterval),
		math.Min(
			trackerRetryDelayMax.Seconds(),
			(trackerRetryDelayMin+failSquare*trackerRetryDelayMin).Seconds()*float64(backoffRatio/100),
		),
	)

	t.trackerState.nextAnnounce = time.Now().Add(time.Duration(delay) + time.Second)
	t.trackerState.updating = false
}

func (t *tracker) Reset() {
	t.trackerState = &trackerState{}
}

func (t *tracker) enqueueAnnounceHistory(announceHistory AnnounceHistory) {
	t.trackerState.announceHistory = append(t.trackerState.announceHistory, announceHistory)

	if len(t.trackerState.announceHistory) > trackerMaxHistorySize {
		t.trackerState.announceHistory = t.trackerState.announceHistory[:trackerMaxHistorySize]
	}
}

type trackerState struct {
	// next announce time
	nextAnnounce time.Time
	// number of consecutive fails
	fails int16
	// true if we already sent the START announce to this tracker
	startSent bool
	// true if we sent an announce to this tracker, and we currently are waiting for an answer
	updating        bool
	announceHistory []AnnounceHistory
}

type AnnounceHistory struct {
	interval time.Duration
	seeders  int32
	leechers int32
	error    string
}
