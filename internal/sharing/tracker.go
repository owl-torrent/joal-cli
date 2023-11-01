package sharing

import (
	"github.com/anthonyraymond/joal-cli/pkg/duration"
	"github.com/google/uuid"
	"net/url"
	"time"
)

type AnnounceEvent uint8

const (
	None      AnnounceEvent = iota // None is a regular announce
	Completed                      // Completed announce sent when the download process is done
	Started                        // Started announce is sent when the torrent start/resume
	Stopped                        // Stopped announce is sent when the torrent is stopped
)

type Tracker struct {
	url              *url.URL
	consecutiveFails int
	nextAnnounceAt   time.Time
	disabled         TrackerDisabled
	hasAnnouncedOnce bool
	pendingAnnounce  *trackerAnnounceRequest // pendingAnnounce store announce sent and waiting for response
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

func (t *Tracker) requireAnnounce(at time.Time) bool {
	return t.pendingAnnounce == nil && t.nextAnnounceAt.Before(at) && !t.isDisabled()
}

func (t *Tracker) announce(event AnnounceEvent) *trackerAnnounceRequest {
	if !t.hasAnnouncedOnce && event == None {
		event = Started
	}

	announceRequest := newAnnounceRequest(t.url, event)
	t.pendingAnnounce = announceRequest

	return announceRequest
}

func (t *Tracker) announceSucceed(response TrackerAnnounceResponse) {
	if !t.isResponseExpected(response.announceUid) {
		return
	}
	t.pendingAnnounce = nil
	t.consecutiveFails = 0
	t.nextAnnounceAt = time.Now().Add(response.Interval)
	t.hasAnnouncedOnce = true
}

func (t *Tracker) announceFailed(error TrackerAnnounceError) {
	if !t.isResponseExpected(error.announceUid) {
		return
	}
	t.pendingAnnounce = nil
	t.nextAnnounceAt = time.Now().Add(calculateBackoff(t.consecutiveFails, 5*time.Second, 1800*time.Second))
	t.consecutiveFails++
}

func (t *Tracker) isResponseExpected(announceUid uuid.UUID) bool {
	if t.pendingAnnounce == nil {
		return false
	}
	return t.pendingAnnounce.uid == announceUid
}

type TrackerAnnounceResponse struct {
	Interval    time.Duration
	announceUid uuid.UUID
}

type TrackerAnnounceError struct {
	announceUid uuid.UUID
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

type trackerAnnounceRequest struct {
	uid   uuid.UUID
	url   *url.URL
	event AnnounceEvent
}

func newAnnounceRequest(u *url.URL, event AnnounceEvent) *trackerAnnounceRequest {
	return &trackerAnnounceRequest{
		uid:   uuid.New(),
		url:   u,
		event: event,
	}
}
