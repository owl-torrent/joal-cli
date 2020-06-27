package torrent

import (
	"context"
	"github.com/anacrolix/torrent/tracker"
	"github.com/google/uuid"
	"net/url"
	"time"
)

var (
	DefaultDurationWaitOnError = 1800 * time.Second
)

type trackerAnnouncer struct {
	uuid         uuid.UUID
	url          url.URL
	responses    chan trackerAwareAnnounceResult
	stoppingLoop chan chan struct{}
}

func newTracker(url url.URL) *trackerAnnouncer {
	return &trackerAnnouncer{
		uuid:         uuid.New(),
		url:          url,
		responses:    make(chan trackerAwareAnnounceResult),
		stoppingLoop: make(chan chan struct{}),
	}
}

func (t trackerAnnouncer) Uuid() uuid.UUID {
	return t.uuid
}

func (t trackerAnnouncer) Responses() <-chan trackerAwareAnnounceResult {
	return t.responses
}

func (t trackerAnnouncer) announceOnce(announce AnnouncingFunction, event tracker.AnnounceEvent) trackerAwareAnnounceResult {
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	return trackerAwareAnnounceResult{
		trackerAnnounceResult: announce(t.url, event, ctx),
		trackerUuid:           t.uuid,
	}
}

func (t *trackerAnnouncer) startAnnounceLoop(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) {
	var next time.Time
	var lastAnnounce trackerAwareAnnounceResult
	event := firstEvent

	var announceDone chan trackerAwareAnnounceResult
	var cancelRunningAnnounce context.CancelFunc
	var pending []trackerAwareAnnounceResult

	for {
		var announceDelay time.Duration
		if now := time.Now(); next.After(now) {
			announceDelay = next.Sub(now)
		}

		// Prevent enqueue another request if the previous one is still on the way
		var announceTime <-chan time.Time
		if announceDone == nil {
			announceTime = time.After(announceDelay)
		}

		// Build some kind of a queue system to ensure the response handling in <- announceDone wont be stuck trying to write to the t.response chan with no one to listen on the other side
		var first trackerAwareAnnounceResult
		var responses chan trackerAwareAnnounceResult
		if len(pending) > 0 {
			first = pending[0]
			responses = t.responses
		}

		select {
		case <-announceTime:
			announceDone = make(chan trackerAwareAnnounceResult, 1)
			go func(t trackerAnnouncer) {
				var ctx context.Context
				ctx, cancelRunningAnnounce = context.WithCancel(context.Background())
				response := announce(t.url, event, ctx)
				announceDone <- trackerAwareAnnounceResult{
					trackerAnnounceResult: response,
					trackerUuid:           t.uuid,
				}
			}(*t)
		case response := <-announceDone:
			cancelRunningAnnounce = nil
			announceDone = nil
			event = tracker.None
			lastAnnounce = response

			var nextAnnounceInterval = response.Interval
			if response.Err != nil {
				nextAnnounceInterval = lastAnnounce.Interval
			}
			if nextAnnounceInterval == 0 {
				nextAnnounceInterval = DefaultDurationWaitOnError
			}
			next = time.Now().Add(nextAnnounceInterval)

			pending = append(pending, response) // enqueue event here and the select will distribute the response as soon as someone is able to read
		case stopDone := <-t.stoppingLoop:
			if cancelRunningAnnounce != nil {
				cancelRunningAnnounce()
			}
			stopDone <- struct{}{}
			return
		case responses <- first:
			pending = pending[1:]
		}
	}
}

func (t *trackerAnnouncer) stopAnnounceLoop() {
	done := make(chan struct{})
	t.stoppingLoop <- done
	<-done
}

type trackerAwareAnnounceResult struct {
	trackerAnnounceResult
	trackerUuid uuid.UUID
}
