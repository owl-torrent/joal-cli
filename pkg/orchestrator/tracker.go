package orchestrator

import (
	"context"
	"errors"
	"github.com/anacrolix/torrent/tracker"
	"net/url"
	"sync"
	"time"
)

const (
	defaultTrackersInterval = 1800 * time.Second
)

type trackerAnnouncer struct {
	url            url.URL
	stoppingLoop   chan chan struct{}
	loopInProgress bool
	lock           *sync.RWMutex
}

func newTracker(url url.URL) ITrackerAnnouncer {
	return &trackerAnnouncer{
		url:            url,
		stoppingLoop:   make(chan chan struct{}),
		loopInProgress: false,
		lock:           &sync.RWMutex{},
	}
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
		Interval:  res.Interval * time.Second,
		Completed: time.Now(),
	}
}

func (t *trackerAnnouncer) startAnnounceLoop(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan trackerAnnounceResult, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.loopInProgress {
		return nil, errors.New("already started")
	}
	t.loopInProgress = true

	var next time.Time
	var lastAnnounce trackerAnnounceResult
	event := firstEvent

	var announceDone chan trackerAnnounceResult
	var cancelRunningAnnounce context.CancelFunc

	responses := make(chan trackerAnnounceResult, 1)

	go func() {
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

			select {
			case <-announceTime:
				announceDone = make(chan trackerAnnounceResult, 1)
				go func(t trackerAnnouncer) {
					var ctx context.Context
					ctx, cancelRunningAnnounce = context.WithCancel(context.Background())

					res, err := announce(ctx, t.url, event)
					if ctx.Err() != nil { // if context is expire do NOT send the event; the context has been canceled = the tracker is stopped = channel is closed
						return
					}
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
						Interval:  res.Interval,
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
					nextAnnounceInterval = defaultTrackersInterval
				}
				next = time.Now().Add(nextAnnounceInterval)

				responses <- response
			case stopDone := <-t.stoppingLoop:
				if cancelRunningAnnounce != nil {
					cancelRunningAnnounce()
				}
				drainResponsesChannel(responses)
				close(responses)
				stopDone <- struct{}{}
				return
			}
		}
	}()

	return responses, nil
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

func drainResponsesChannel(c <-chan trackerAnnounceResult) {
	for {
		select {
		case <-c:
			continue
		default:
			return
		}
	}
}
