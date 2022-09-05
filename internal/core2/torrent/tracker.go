package torrent

import (
	"math"
	"net/url"
	"time"
)

const (
	backoffRatio          = float64(250)
	minimumRetryDelay     = 5 * time.Second
	maximumRetryDelay     = 60 * time.Minute
	announceHistoryMaxLen = 5
)

var (
	announceProtocolNotSupported = TrackerDisabled{disabled: true, reason: "client.announce.protocol.not-supported"}
	announceListNotSupported     = TrackerDisabled{disabled: true, reason: "client.announce.announce-list.not-supported"}
)

type Tracker struct {
	url                   url.URL
	tier                  int
	disabled              TrackerDisabled
	nextAnnounce          time.Time
	consecutiveFails      int
	startSent             bool
	isCurrentlyAnnouncing bool
	announcesHistory      []AnnounceHistory
}

func (t *Tracker) getTier() int {
	return t.tier
}

func (t *Tracker) isAnnouncing() bool {
	return t.isCurrentlyAnnouncing
}

// return true if the tracker is the tracker is ready to announce
// (not disabled && nextAnnounce is passed && not announcing)
func (t *Tracker) canAnnounce() bool {
	if t.disabled.IsDisabled() {
		return false
	}
	if time.Now().Before(t.nextAnnounce) {
		return false
	}
	if t.isCurrentlyAnnouncing {
		return false
	}
	return true
}

func (t *Tracker) HasAnnouncedStart() bool {
	return t.startSent
}

func (t *Tracker) announcing() {
	t.isCurrentlyAnnouncing = true
}

func (t *Tracker) announceSucceed(h AnnounceHistory) {
	t.consecutiveFails = 0
	t.isCurrentlyAnnouncing = false
	t.announcesHistory = appendToAnnounceHistory(t.announcesHistory, h, announceHistoryMaxLen)

	t.nextAnnounce = h.at.Add(h.interval)
}

func (t *Tracker) announceFailed(h AnnounceHistory) {
	if t.consecutiveFails != math.MaxInt {
		t.consecutiveFails = t.consecutiveFails + 1
	}
	t.isCurrentlyAnnouncing = false
	t.announcesHistory = appendToAnnounceHistory(t.announcesHistory, h, announceHistoryMaxLen)

	failSquare := float64(t.consecutiveFails * t.consecutiveFails)

	// the exponential back-off ends up being:
	// 17, 55, 117, 205, 317, 455, 617, ... seconds
	// with tracker_backoff = 250
	delay := math.Max(
		h.interval.Seconds(),
		math.Min(
			maximumRetryDelay.Seconds(),
			minimumRetryDelay.Seconds()+failSquare*minimumRetryDelay.Seconds()*backoffRatio/float64(100),
		),
	)

	t.nextAnnounce = h.at.Add(time.Duration(delay) * time.Second)
}

func appendToAnnounceHistory(slice []AnnounceHistory, h AnnounceHistory, maxLen int) []AnnounceHistory {
	slice = append(slice, h)

	if len(slice) > maxLen {
		slice = slice[:maxLen]
	}
	return slice
}

type AnnouncePolicy interface {
	SupportHttpAnnounce() bool
	SupportUdpAnnounce() bool
	ShouldSupportAnnounceList() bool
	ShouldAnnounceToAllTier() bool
	ShouldAnnounceToAllTrackersInTier() bool
}

type TrackerDisabled struct {
	disabled bool
	reason   string
}

func (d TrackerDisabled) IsDisabled() bool {
	return d.disabled
}

type AnnounceHistory struct {
	at       time.Time
	interval time.Duration
	seeders  int32
	leechers int32
	error    string
}
