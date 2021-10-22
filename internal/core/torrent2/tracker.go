package torrent2

import (
	"github.com/anacrolix/torrent/metainfo"
	"math"
	"math/rand"
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
	Url() url.URL
	Tier() int16
	Disable()
	CanAnnounce(now time.Time) bool
	IsWorking() bool
	Succeed(announceHistory AnnounceHistory)
	Failed(announceHistory AnnounceHistory, backoffRatio int, retryInterval int)
	Reset()
}

type tracker struct {
	url *url.URL
	// tier that the tracker belongs to
	tier int16
	// if false the tracker must not be used at all
	enabled      bool
	trackerState *trackerState
}

func newTrackers(announce string, announceList metainfo.AnnounceList, supportAnnounceList bool) []Tracker {
	// Shuffling trackers according to BEP-12: https://www.bittorrent.org/beps/bep_0012.html
	rand.Seed(randSeed)
	for _, tier := range announceList {
		rand.Shuffle(len(tier), func(i, j int) {
			tier[i], tier[j] = tier[j], tier[i]
		})
	}

	// tracker does not support annnouceList OR AnnounceList contains only empty url but announce contains a valid url
	if !supportAnnounceList || !announceList.OverridesAnnounce(announce) {
		u, err := url.Parse(announce)
		if err != nil {
			return []Tracker{}
		}
		return []Tracker{newTracker(u, 0)}
	}

	var trackers []Tracker
	for tierIndex, tier := range announceList {
		for _, trackerUri := range tier {
			u, err := url.Parse(trackerUri)
			if err != nil {
				continue
			}
			trackers = append(trackers, newTracker(u, int16(tierIndex)))
		}
	}
	return trackers
}

func newTracker(u *url.URL, tierIndex int16) Tracker {
	return &tracker{
		url:          u,
		tier:         tierIndex,
		enabled:      true,
		trackerState: &trackerState{},
	}
}

func (t *tracker) Url() url.URL {
	return *t.url
}
func (t *tracker) Tier() int16 {
	return t.tier
}

func (t *tracker) Disable() {
	t.enabled = false
}

// CanAnnounce returns true if the tracker is not currently announcing and if the interval to wait is elapsed
func (t *tracker) CanAnnounce(now time.Time) bool {
	if !t.enabled {
		return false
	}
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
