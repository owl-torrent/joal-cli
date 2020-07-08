package orchestrator

import (
	"errors"
	"github.com/anacrolix/torrent/tracker"
	"sync"
	"time"
)

type trackerAwareAnnounceResult struct {
	trackerAnnounceResult
	tracker ITrackerAnnouncer
}

type AllTrackersTierAnnouncer struct {
	trackers          []ITrackerAnnouncer
	state             chan tierState
	stoppingTier      chan chan struct{}
	loopInProgress    bool
	lock              *sync.RWMutex
	lastKnownInterval time.Duration
}

func (t AllTrackersTierAnnouncer) LastKnownInterval() (time.Duration, error) {
	if t.lastKnownInterval == 0 {
		return 0 * time.Nanosecond, errors.New("no interval received from trackers yet")
	}
	return t.lastKnownInterval, nil
}

func newAllTrackersTierAnnouncer(trackers ...ITrackerAnnouncer) (ITierAnnouncer, error) {
	if len(trackers) == 0 {
		return nil, errors.New("a tier can not have an empty tracker list")
	}
	t := &AllTrackersTierAnnouncer{
		trackers:          trackers,
		state:             make(chan tierState),
		stoppingTier:      make(chan chan struct{}),
		loopInProgress:    false,
		lock:              &sync.RWMutex{},
		lastKnownInterval: 0 * time.Nanosecond,
	}

	return t, nil
}

func (t AllTrackersTierAnnouncer) announceOnce(announce AnnouncingFunction, event tracker.AnnounceEvent) tierState {
	wg := sync.WaitGroup{}

	lock := &sync.Mutex{}
	states := make(map[ITrackerAnnouncer]tierState)

	for _, tr := range t.trackers {
		wg.Add(1)
		go func(tr ITrackerAnnouncer) {
			defer wg.Done()
			resp := tr.announceOnce(announce, event)
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

func (t AllTrackersTierAnnouncer) States() <-chan tierState {
	return t.state
}

func (t *AllTrackersTierAnnouncer) startAnnounceLoop(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) {
	t.lock.Lock()
	if t.loopInProgress {
		t.lock.Unlock()
		return
	}
	t.loopInProgress = true
	t.lock.Unlock()

	for _, tr := range t.trackers {
		go tr.startAnnounceLoop(announce, firstEvent)
	}

	responseReceived := make(chan trackerAwareAnnounceResult, len(t.trackers))
	stoppingLoops := make(chan chan struct{}, len(t.trackers))

	for _, tr := range t.trackers {
		go func(tr ITrackerAnnouncer) {
			for {
				select {
				case resp := <-tr.Responses():
					responseReceived <- trackerAwareAnnounceResult{trackerAnnounceResult: resp, tracker: tr}
				case doneStopping := <-stoppingLoops:
					tr.stopAnnounceLoop()
					doneStopping <- struct{}{}
					return
				}
			}
		}(tr)
	}

	trackersStates := make(map[ITrackerAnnouncer]tierState, len(t.trackers))
	// All trackers starts alive
	for _, tr := range t.trackers {
		trackersStates[tr] = ALIVE
	}
	currentTierState := ALIVE

	// this chan will be allocated once every time the tier changes his state to prevent spamming the receiver with the same message at every turn of fhe for loop.
	var stateUpdated chan tierState = nil
	firstStateReported := false // the default state is ALIVE, but we need to report that the tracker is ALIVE on the first success or DEAD after all failed

	for {
		select {
		case resp := <-responseReceived:
			var ts trackerState = DEAD
			if resp.Err == nil {
				ts = ALIVE
				t.lastKnownInterval = resp.Interval
			}
			trackersStates[resp.tracker] = ts
			if !firstStateReported && ts == ALIVE {
				currentTierState = ALIVE
				stateUpdated = t.state
				continue
			}

			if ts == currentTierState {
				continue
			}

			// recalculate the tier state with the new information available
			var stateAfterUpdate tierState = DEAD
			for _, trs := range trackersStates {
				if trs == ALIVE {
					stateAfterUpdate = ALIVE
					break
				}
			}
			if stateAfterUpdate == currentTierState {
				continue
			}

			currentTierState = stateAfterUpdate
			stateUpdated = t.state
		case stateUpdated <- currentTierState:
			firstStateReported = true
			stateUpdated = nil
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
						case <-time.After(25 * time.Millisecond):
							drainTierResponseChannel(responseReceived) // drain channel ensuring no goroutine are blocked writting to it
						}
					}

				}()
			}
			wg.Wait()
			doneStopping <- struct{}{}

			return
		}
	}
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
	tracker           *linkedTrackerList
	state             chan tierState
	stoppingTier      chan chan struct{}
	loopInProgress    bool
	lock              *sync.RWMutex
	lastKnownInterval time.Duration
}

func newFallbackTrackersTierAnnouncer(trackers ...ITrackerAnnouncer) (ITierAnnouncer, error) {
	if len(trackers) == 0 {
		return nil, errors.New("a tier can not have an empty tracker list")
	}
	list, err := newLinkedTrackerList(trackers)
	if err != nil {
		return nil, err
	}
	t := &FallbackTrackersTierAnnouncer{
		tracker:           list,
		state:             make(chan tierState),
		stoppingTier:      make(chan chan struct{}),
		loopInProgress:    false,
		lock:              &sync.RWMutex{},
		lastKnownInterval: 0 * time.Nanosecond,
	}

	return t, nil
}

func (t FallbackTrackersTierAnnouncer) LastKnownInterval() (time.Duration, error) {
	if t.lastKnownInterval == 0 {
		return 0 * time.Nanosecond, errors.New("no interval received from trackers yet")
	}
	return t.lastKnownInterval, nil
}

func (t *FallbackTrackersTierAnnouncer) announceOnce(announce AnnouncingFunction, event tracker.AnnounceEvent) tierState {
	res := t.tracker.announceOnce(announce, event)
	if res.Err == nil {
		t.tracker.PromoteCurrent()
		return ALIVE
	}

	for !t.tracker.isLast() {
		t.tracker.next()
		res := t.tracker.announceOnce(announce, event)
		if res.Err == nil {
			t.tracker.PromoteCurrent()
			return ALIVE
		}
	}

	return tierState(DEAD)
}

func (t FallbackTrackersTierAnnouncer) States() <-chan tierState {
	return t.state
}

func (t *FallbackTrackersTierAnnouncer) startAnnounceLoop(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) {
	t.lock.Lock()
	if t.loopInProgress {
		t.lock.Unlock()
		return
	}
	t.loopInProgress = true
	t.lock.Unlock()

	stoppingLoop := make(chan chan struct{})
	pauseBeforeLoop := time.After(0 * time.Millisecond)
	responseReceived := make(chan trackerAwareAnnounceResult, 5) // 5 because why not... messages should not enqueue too much since all tracker are announcing one by one anyway

	currentEvent := firstEvent

	go func() {
		for {
			select {
			case res := <-t.tracker.Responses():
				pauseBeforeLoop = nil
				responseReceived <- trackerAwareAnnounceResult{trackerAnnounceResult: res, tracker: t.tracker.ITrackerAnnouncer}
				if res.Err == nil {
					t.lastKnownInterval = res.Interval
					currentEvent = tracker.None
					t.tracker.PromoteCurrent()
					continue
				}

				t.tracker.stopAnnounceLoop()
				t.tracker.next()
				if !t.tracker.isFirst() {
					pauseBeforeLoop = time.After(0 * time.Millisecond)
					continue
				}

				// reached the end of the list
				pause := t.lastKnownInterval
				if t.lastKnownInterval == 0 {
					pause = DefaultDurationWaitOnError
				}
				pauseBeforeLoop = time.After(pause)

			case <-pauseBeforeLoop:
				go t.tracker.startAnnounceLoop(announce, currentEvent)
			case doneStopping := <-stoppingLoop:
				t.tracker.stopAnnounceLoop()
				doneStopping <- struct{}{}
				return
			}
		}
	}()

	trackersStates := make(map[ITrackerAnnouncer]tierState, len(t.tracker.list))
	// All trackers starts alive
	for _, tr := range t.tracker.list {
		trackersStates[tr] = ALIVE
	}
	currentTierState := ALIVE

	// this chan will be allocated once every time the tier changes his state to prevent spamming the receiver with the same message at every turn of fhe for loop.
	var stateUpdated chan tierState = nil
	firstStateReported := false // the default state is ALIVE, but we need to report that the tracker is ALIVE on the first success or DEAD after all failed

	for {
		select {
		case res := <-responseReceived:
			var ts trackerState = DEAD
			if res.Err == nil {
				ts = ALIVE
			}
			trackersStates[res.tracker] = ts
			if !firstStateReported && ts == ALIVE {
				currentTierState = ALIVE
				stateUpdated = t.state
				continue
			}

			if ts == currentTierState {
				continue
			}

			// recalculate the tier state with the new information available
			var stateAfterUpdate tierState = DEAD
			for _, trs := range trackersStates {
				if trs == ALIVE {
					stateAfterUpdate = ALIVE
					break
				}
			}
			if stateAfterUpdate == currentTierState {
				continue
			}

			currentTierState = stateAfterUpdate
			stateUpdated = t.state
		case stateUpdated <- currentTierState:
			firstStateReported = true
			stateUpdated = nil
		case doneStopping := <-t.stoppingTier:
			done := make(chan struct{})
			stoppingLoop <- done

			for {
				select {
				case <-done:
					doneStopping <- struct{}{}

					return
				case <-time.After(25 * time.Millisecond):
					drainTierResponseChannel(responseReceived) // drain channel ensuring no goroutine are blocked writting to it
				}
			}
		}
	}
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

func drainTierResponseChannel(t <-chan trackerAwareAnnounceResult) {
	for {
		select {
		case <-t:
			continue
		default:
			return
		}
	}
}
