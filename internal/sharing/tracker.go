package sharing

import (
	"fmt"
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

type Tracker interface {
	disable(reason TrackerDisableReason)
	requireAnnounce(at time.Time) bool
	announce(event AnnounceEvent) (*trackerAnnounceRequest, error)
	announceSucceed(response TrackerAnnounceResponse) error
	announceFailed(error TrackerAnnounceError) error
	Url() *url.URL
	ConsecutiveFails() int
	IsDisabled() bool
	NextAnnounceAt() time.Time
}

type trackerImpl struct {
	url                  *url.URL
	consecutiveFailCount int
	nextAnnounceAt       time.Time
	disabled             TrackerDisabled
	hasAnnouncedOnce     bool
	pendingAnnounce      *trackerAnnounceRequest // pendingAnnounce store announce sent and waiting for response
}

func newTracker(u *url.URL) Tracker {
	return &trackerImpl{
		url:             u,
		pendingAnnounce: nil,
	}
}

func (t *trackerImpl) Url() *url.URL {
	return t.url
}

func (t *trackerImpl) ConsecutiveFails() int {
	return t.consecutiveFailCount
}

func (t *trackerImpl) IsDisabled() bool {
	return t.disabled.isDisabled
}

func (t *trackerImpl) NextAnnounceAt() time.Time {
	return t.nextAnnounceAt
}

func (t *trackerImpl) disable(reason TrackerDisableReason) {
	t.disabled = TrackerDisabled{
		isDisabled: true,
		reason:     reason,
	}
}

func (t *trackerImpl) requireAnnounce(at time.Time) bool {
	return t.pendingAnnounce == nil && t.nextAnnounceAt.Before(at) && !t.IsDisabled()
}

func (t *trackerImpl) announce(event AnnounceEvent) (*trackerAnnounceRequest, error) {
	if t.pendingAnnounce != nil && event != Stopped && event != Completed {
		return nil, fmt.Errorf("awaiting for an announce response, can't send a concurent [%s] announce", string(event))
	}

	if !t.hasAnnouncedOnce && event == None {
		event = Started
	}

	announceRequest := newAnnounceRequest(t.url, event)
	t.pendingAnnounce = announceRequest

	return announceRequest, nil
}

func (t *trackerImpl) announceSucceed(response TrackerAnnounceResponse) error {
	if !t.isResponseExpected(response.announceUid) {
		return fmt.Errorf("unexpected announce response [%s]", response.announceUid.String())
	}
	t.pendingAnnounce = nil
	t.consecutiveFailCount = 0
	t.nextAnnounceAt = time.Now().Add(response.Interval)
	t.hasAnnouncedOnce = true
	return nil
}

func (t *trackerImpl) announceFailed(error TrackerAnnounceError) error {
	if !t.isResponseExpected(error.announceUid) {
		return fmt.Errorf("unexpected announce response [%s]", error.announceUid.String())
	}
	t.pendingAnnounce = nil
	t.nextAnnounceAt = time.Now().Add(calculateBackoff(t.consecutiveFailCount, 5*time.Second, 1800*time.Second))
	t.consecutiveFailCount++
	return nil
}

func (t *trackerImpl) isResponseExpected(announceUid uuid.UUID) bool {
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
