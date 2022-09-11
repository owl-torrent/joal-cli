package torrent

import (
	"fmt"
	trackerlib "github.com/anacrolix/torrent/tracker"
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
	announceProtocolNotSupported = trackerDisabled{disabled: true, reason: "tracker.disabled.protocol-not-supported"}
	announceListNotSupported     = trackerDisabled{disabled: true, reason: "tracker.disabled.announce-list-not-supported"}
)

type tracker struct {
	url                   url.URL
	tier                  int
	disabled              trackerDisabled
	nextAnnounce          time.Time
	consecutiveFails      int
	startSent             bool
	isCurrentlyAnnouncing bool
	announcesHistory      []announceHistory
}

func (t *tracker) getTier() int {
	return t.tier
}

func (t *tracker) isAnnouncing() bool {
	return t.isCurrentlyAnnouncing
}

// return true if the tracker is ready to announce
// (not disabled && nextAnnounce is passed && not announcing)
func (t *tracker) canAnnounce(at time.Time) bool {
	if t.disabled.isDisabled() {
		return false
	}
	if at.Before(t.nextAnnounce) {
		return false
	}
	if t.isCurrentlyAnnouncing {
		return false
	}
	return true
}

func (t *tracker) announce(event trackerlib.AnnounceEvent, contrib contribution, announceFunction AnnouncingFunction) error {
	if t.disabled.disabled {
		return fmt.Errorf("can not announce, tracker is disabled")
	}

	// if we never announced start replace the None announce with a start
	if event == trackerlib.None && !t.startSent {
		event = trackerlib.Started
	}
	// if we never announced start, there is no need to announce Stopped
	if event == trackerlib.Stopped && !t.startSent {
		return nil
	}

	t.isCurrentlyAnnouncing = true
	announceFunction(TrackerAnnounceRequest{
		Event:      event,
		Url:        t.url,
		Uploaded:   contrib.uploaded,
		Downloaded: contrib.downloaded,
		Left:       contrib.left,
		Corrupt:    contrib.corrupt,
	})

	return nil
}

func (t *tracker) announceSucceed(h announceHistory) {
	t.consecutiveFails = 0
	t.isCurrentlyAnnouncing = false
	t.announcesHistory = appendToAnnounceHistory(t.announcesHistory, h, announceHistoryMaxLen)

	t.nextAnnounce = h.at.Add(h.interval)
}

func (t *tracker) announceFailed(h announceHistory) {
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

func appendToAnnounceHistory(slice []announceHistory, h announceHistory, maxLen int) []announceHistory {
	slice = append(slice, h)

	if len(slice) > maxLen {
		slice = slice[:maxLen]
	}
	return slice
}

type trackerDisabled struct {
	disabled bool
	reason   string
}

func (d trackerDisabled) isDisabled() bool {
	return d.disabled
}

type announceHistory struct {
	at       time.Time
	interval time.Duration
	seeders  int32
	leechers int32
	error    string
}
