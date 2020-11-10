package orchestrator

import (
	"context"
	"fmt"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/logs"
	"go.uber.org/zap"
	"sync"
	"time"
)

type AllTrackersTierAnnouncer struct {
	trackers          []ITrackerAnnouncer
	state             chan tierState
	stoppingTier      chan chan struct{}
	loopInProgress    bool
	lock              *sync.RWMutex
	lastKnownInterval time.Duration
}

func (t AllTrackersTierAnnouncer) LastKnownInterval() time.Duration {
	return t.lastKnownInterval
}

func newAllTrackersTierAnnouncer(trackers ...ITrackerAnnouncer) (ITierAnnouncer, error) {
	if len(trackers) == 0 {
		return nil, fmt.Errorf("a tier can not have an empty tracker list")
	}
	t := &AllTrackersTierAnnouncer{
		trackers:          trackers,
		state:             make(chan tierState),
		stoppingTier:      make(chan chan struct{}),
		loopInProgress:    false,
		lock:              &sync.RWMutex{},
		lastKnownInterval: defaultTrackersInterval,
	}

	return t, nil
}

func (t AllTrackersTierAnnouncer) announceOnce(ctx context.Context, announce AnnouncingFunction, event tracker.AnnounceEvent) tierState {
	select {
	case <-ctx.Done():
		return DEAD
	default:
	}
	wg := sync.WaitGroup{}

	lock := &sync.Mutex{}
	states := make(map[ITrackerAnnouncer]tierState)

	for _, tr := range t.trackers {
		wg.Add(1)
		go func(tr ITrackerAnnouncer) {
			defer wg.Done()
			resp := tr.announceOnce(ctx, announce, event)
			lock.Lock()
			st := ALIVE
			if resp.Err != nil {
				st = DEAD
			}
			states[tr] = st
			lock.Unlock()
		}(tr)
	}

	wg.Wait()

	for _, st := range states {
		if st == ALIVE {
			return ALIVE
		}
	}
	return DEAD
}

func (t *AllTrackersTierAnnouncer) startAnnounceLoop(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.loopInProgress {
		return nil, fmt.Errorf("already started")
	}
	t.loopInProgress = true

	stateCalculator := NewTierStateCalculator(t.trackers)

	stoppingLoops := make(chan chan struct{}, len(t.trackers))

	for _, tr := range t.trackers {
		go func(tr ITrackerAnnouncer) {
			trackerResponses, err := tr.startAnnounceLoop(announce, firstEvent)
			if err != nil {
				tr.stopAnnounceLoop()
				logs.GetLogger().Warn("failed to start announcing", zap.Error(err))
			}
			for {
				select {
				case resp := <-trackerResponses:
					if resp.Err != nil {
						stateCalculator.setIndividualState(tr, false)
						continue
					}
					stateCalculator.setIndividualState(tr, true)
					t.lastKnownInterval = resp.Interval
				case doneStopping := <-stoppingLoops:
					tr.stopAnnounceLoop()
					doneStopping <- struct{}{}
					return
				}
			}
		}(tr)
	}

	go func() {
		select {
		case doneStopping := <-t.stoppingTier:
			wg := sync.WaitGroup{}

			for range t.trackers {
				wg.Add(1)
				go func() {
					defer wg.Done()
					done := make(chan struct{})
					stoppingLoops <- done
					for {
						select {
						case <-done:
							return
						}
					}

				}()
			}
			wg.Wait()
			stateCalculator.Destroy()
			doneStopping <- struct{}{}
		}
	}()

	return stateCalculator.States(), nil
}

func (t *AllTrackersTierAnnouncer) stopAnnounceLoop() {
	t.lock.Lock()
	defer t.lock.Unlock()
	if !t.loopInProgress {
		return
	}
	t.loopInProgress = false

	done := make(chan struct{})
	t.stoppingTier <- done
	<-done
}

type FallbackTrackersTierAnnouncer struct {
	allTrackers       []ITrackerAnnouncer
	tracker           *linkedTrackerList
	state             chan tierState
	stoppingTier      chan chan struct{}
	loopInProgress    bool
	lock              *sync.RWMutex
	lastKnownInterval time.Duration
}

func newFallbackTrackersTierAnnouncer(trackers ...ITrackerAnnouncer) (ITierAnnouncer, error) {
	if len(trackers) == 0 {
		return nil, fmt.Errorf("a tier can not have an empty tracker list")
	}
	list, err := newLinkedTrackerList(trackers)
	if err != nil {
		return nil, err
	}
	t := &FallbackTrackersTierAnnouncer{
		allTrackers:       trackers,
		tracker:           list,
		state:             make(chan tierState),
		stoppingTier:      make(chan chan struct{}),
		loopInProgress:    false,
		lock:              &sync.RWMutex{},
		lastKnownInterval: defaultTrackersInterval,
	}

	return t, nil
}

func (t FallbackTrackersTierAnnouncer) LastKnownInterval() time.Duration {
	return t.lastKnownInterval
}

func (t *FallbackTrackersTierAnnouncer) announceOnce(ctx context.Context, announce AnnouncingFunction, event tracker.AnnounceEvent) tierState {
	select {
	case <-ctx.Done():
		return DEAD
	default:
	}
	res := t.tracker.announceOnce(ctx, announce, event)
	if res.Err == nil {
		t.tracker.PromoteCurrent()
		return ALIVE
	}

	select {
	case <-ctx.Done():
		return DEAD
	default:
	}

	for !t.tracker.isLast() {
		t.tracker.next()
		res := t.tracker.announceOnce(ctx, announce, event)
		if res.Err == nil {
			t.tracker.PromoteCurrent()
			return ALIVE
		}
	}

	return DEAD
}

func (t FallbackTrackersTierAnnouncer) States() <-chan tierState {
	return t.state
}

func (t *FallbackTrackersTierAnnouncer) startAnnounceLoop(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.loopInProgress {
		return nil, fmt.Errorf("already started")
	}
	t.loopInProgress = true

	stateCalculator := NewTierStateCalculator(t.allTrackers)

	pauseBeforeLoop := time.After(0 * time.Millisecond)
	currentEvent := firstEvent
	var responsesChan <-chan trackerAnnounceResult

	go func() {
		for {
			select {
			case <-pauseBeforeLoop:
				var err error = nil
				responsesChan, err = t.tracker.startAnnounceLoop(announce, currentEvent)
				if err != nil {
					t.tracker.stopAnnounceLoop()
					responsesChan = nil
					t.tracker.next()
					pauseBeforeLoop = time.After(5 * time.Second)
				}
			case res := <-responsesChan:
				pauseBeforeLoop = nil
				if res.Err == nil {
					stateCalculator.setIndividualState(t.tracker.ITrackerAnnouncer, true)
					t.lastKnownInterval = res.Interval
					currentEvent = tracker.None
					t.tracker.PromoteCurrent()
					continue
				}
				stateCalculator.setIndividualState(t.tracker.ITrackerAnnouncer, false)

				t.tracker.stopAnnounceLoop()
				responsesChan = nil
				t.tracker.next()
				if !t.tracker.isFirst() {
					pauseBeforeLoop = time.After(0 * time.Millisecond)
					continue
				}

				// reached the end of the list
				pauseBeforeLoop = time.After(t.lastKnownInterval)
			case doneStopping := <-t.stoppingTier:
				t.tracker.stopAnnounceLoop()
				stateCalculator.Destroy()
				responsesChan = nil
				doneStopping <- struct{}{}
				return
			}
		}
	}()

	return stateCalculator.States(), nil
}

func (t *FallbackTrackersTierAnnouncer) stopAnnounceLoop() {
	t.lock.Lock()
	defer t.lock.Unlock()
	if !t.loopInProgress {
		return
	}
	t.loopInProgress = false

	done := make(chan struct{})
	t.stoppingTier <- done
	<-done
}
