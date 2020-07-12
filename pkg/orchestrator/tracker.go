package orchestrator

import (
	"context"
	"github.com/anacrolix/torrent/tracker"
	"net/url"
	"sync"
	"time"
)

var (
	DefaultDurationWaitOnError = 1800 * time.Second
)

type trackerAnnouncer struct {
	url            url.URL
	responses      chan trackerAnnounceResult
	stoppingLoop   chan chan struct{}
	loopInProgress bool
	lock           *sync.RWMutex
}

func newTracker(url url.URL) *trackerAnnouncer {
	return &trackerAnnouncer{
		url:            url,
		responses:      make(chan trackerAnnounceResult),
		stoppingLoop:   make(chan chan struct{}),
		loopInProgress: false,
		lock:           &sync.RWMutex{},
	}
}

func (t trackerAnnouncer) Responses() <-chan trackerAnnounceResult {
	return t.responses
}

func (t trackerAnnouncer) announceOnce(ctx context.Context, announce AnnouncingFunction, event tracker.AnnounceEvent) trackerAnnounceResult {
	res, err := announce(ctx, t.url, event)
	if err != nil {
		return trackerAnnounceResult{
			Err:       err,
			Interval:  0,
			Completed: time.Now(),
		}
	}
	return trackerAnnounceResult{
		Err:       nil,
		Interval:  time.Duration(res.Interval) * time.Second,
		Completed: time.Now(),
	}
}

func (t *trackerAnnouncer) startAnnounceLoop(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) {
	t.lock.Lock()
	if t.loopInProgress {
		t.lock.Unlock()
		return
	}
	t.loopInProgress = true
	t.lock.Unlock()

	var next time.Time
	var lastAnnounce trackerAnnounceResult
	event := firstEvent

	var announceDone chan trackerAnnounceResult
	var cancelRunningAnnounce context.CancelFunc
	var pendingResponses []trackerAnnounceResult

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
		var firstPendingResponse trackerAnnounceResult
		var responses chan trackerAnnounceResult
		if len(pendingResponses) > 0 {
			firstPendingResponse = pendingResponses[0]
			responses = t.responses
		}

		select {
		case <-announceTime:
			announceDone = make(chan trackerAnnounceResult, 1)
			go func(t trackerAnnouncer) {
				var ctx context.Context
				ctx, cancelRunningAnnounce = context.WithCancel(context.Background())

				res, err := announce(ctx, t.url, event)
				if err != nil {
					announceDone <- trackerAnnounceResult{
						Err:       err,
						Interval:  0,
						Completed: time.Now(),
					}
					return
				}
				announceDone <- trackerAnnounceResult{
					Err:       nil,
					Interval:  time.Duration(res.Interval) * time.Second,
					Completed: time.Now(),
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

			pendingResponses = append(pendingResponses, response) // enqueue event here and the select will distribute the response as soon as someone is able to read
		case stopDone := <-t.stoppingLoop:
			if cancelRunningAnnounce != nil {
				cancelRunningAnnounce()
			}
			stopDone <- struct{}{}
			return
		case responses <- firstPendingResponse:
			pendingResponses = pendingResponses[1:]
		}
	}
}

func (t *trackerAnnouncer) stopAnnounceLoop() {
	t.lock.Lock()
	defer t.lock.Unlock()
	if !t.loopInProgress {
		return
	}
	t.loopInProgress = false

	done := make(chan struct{})
	t.stoppingLoop <- done
	<-done
}
