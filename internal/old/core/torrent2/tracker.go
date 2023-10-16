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

// Maximum number of announce history to keep for each trackerImpl
const trackerMaxHistorySize = 3

type Tracker interface {
	Url() url.URL
	CanAnnounce(now time.Time) bool
	IsWorking() bool
	Succeed(announceHistory AnnounceHistory)
	Failed(announceHistory AnnounceHistory, backoffRatio int, retryInterval int)
	Reset()
}

type trackerImpl struct {
	url *url.URL
	// tier that the trackerImpl belongs to
	tier int16
	// if false the trackerImpl must not be used at all
	enabled bool
	state   *trackerState
}

func newTrackers(announce string, announceList metainfo.AnnounceList, supportAnnounceList bool) []*trackerImpl {
	// Shuffling trackers according to BEP-12: https://www.bittorrent.org/beps/bep_0012.html
	rand.Seed(randSeed)
	for _, tier := range announceList {
		rand.Shuffle(len(tier), func(i, j int) {
			tier[i], tier[j] = tier[j], tier[i]
		})
	}

	// trackerImpl does not support annnouceList OR AnnounceList contains only empty url but announce contains a valid url
	if !supportAnnounceList || !announceList.OverridesAnnounce(announce) {
		u, err := url.Parse(announce)
		if err != nil {
			return []*trackerImpl{}
		}
		return []*trackerImpl{newTracker(u, 0)}
	}

	var trackers []*trackerImpl
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

func newTracker(u *url.URL, tierIndex int16) *trackerImpl {
	return &trackerImpl{
		url:     u,
		tier:    tierIndex,
		enabled: true,
		state:   &trackerState{},
	}
}

func (t *trackerImpl) Url() url.URL {
	return *t.url
}

// CanAnnounce returns true if the trackerImpl is not currently announcing and if the interval to wait is elapsed
func (t *trackerImpl) CanAnnounce(now time.Time) bool {
	if !t.enabled {
		return false
	}
	// add 1 safety sec before comparing "strictly greater than", because
	if now.Before(t.state.nextAnnounce) {
		return false
	}
	if t.state.updating {
		return false
	}
	return true
}

// IsWorking returns true if the last announce was successful
func (t *trackerImpl) IsWorking() bool {
	return t.enabled && t.state.fails == 0
}

func (t *trackerImpl) Succeed(announceHistory AnnounceHistory) {
	t.state.fails = 0
	t.enqueueAnnounceHistory(announceHistory)

	t.state.nextAnnounce = time.Now().Add(announceHistory.interval)

	t.state.updating = false
}

func (t *trackerImpl) Failed(announceHistory AnnounceHistory, backoffRatio int, retryInterval int) {
	t.state.fails++
	t.enqueueAnnounceHistory(announceHistory)

	// the exponential back-off ends up being:
	// 7, 15, 27, 45, 95, 127, 165, ... seconds
	// with the default tracker_backoff of 250
	failSquare := time.Duration(t.state.fails*t.state.fails) * time.Second

	delay := math.Max(
		float64(retryInterval),
		math.Min(
			trackerRetryDelayMax.Seconds(),
			(trackerRetryDelayMin+failSquare*trackerRetryDelayMin).Seconds()*float64(backoffRatio/100),
		),
	)

	t.state.nextAnnounce = time.Now().Add(time.Duration(delay) + time.Second)
	t.state.updating = false
}

func (t *trackerImpl) Reset() {
	t.state = &trackerState{}
}

func (t *trackerImpl) enqueueAnnounceHistory(announceHistory AnnounceHistory) {
	t.state.announceHistory = append(t.state.announceHistory, announceHistory)

	if len(t.state.announceHistory) > trackerMaxHistorySize {
		t.state.announceHistory = t.state.announceHistory[:trackerMaxHistorySize]
	}
}

type trackerState struct {
	// next announce time
	nextAnnounce time.Time
	// number of consecutive fails
	fails int16
	// true if we already sent the START announce to this trackerImpl
	startSent bool
	// true if we sent an announce to this trackerImpl, and we currently are waiting for an answer
	updating        bool
	announceHistory []AnnounceHistory
}

type AnnounceHistory struct {
	interval time.Duration
	seeders  int32
	leechers int32
	error    string
}
