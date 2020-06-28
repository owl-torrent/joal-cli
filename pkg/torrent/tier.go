package torrent

import (
	"errors"
	"github.com/anacrolix/torrent/tracker"
	"github.com/google/uuid"
	"sync"
)

type AllTrackersTierAnnouncer struct {
	uuid          uuid.UUID
	trackers      []ITrackerAnnouncer
	state         chan tierState
	stoppingLoops chan chan struct{}
	stoppingTier  chan chan struct{}
}

func newAllTrackersTierAnnouncer(trackers ...ITrackerAnnouncer) (ITierAnnouncer, error) {
	if len(trackers) == 0 {
		return nil, errors.New("a tier can not have an empty tracker list")
	}
	t := &AllTrackersTierAnnouncer{
		uuid:          uuid.New(),
		trackers:      trackers,
		state:         make(chan tierState),
		stoppingLoops: make(chan chan struct{}),
		stoppingTier:  make(chan chan struct{}),
	}

	return t, nil
}

func (t AllTrackersTierAnnouncer) Uuid() uuid.UUID {
	return t.uuid
}

func (t AllTrackersTierAnnouncer) announceOnce(announce AnnouncingFunction, event tracker.AnnounceEvent) tierState {
	wg := sync.WaitGroup{}

	states := make(map[uuid.UUID]tierState)

	for _, tr := range t.trackers {
		wg.Add(1)
		go func(tr ITrackerAnnouncer) {
			defer wg.Done()
			resp := tr.announceOnce(announce, event)
			if resp.Err != nil {
				states[resp.trackerUuid] = DEAD
			}
			states[resp.trackerUuid] = ALIVE
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
	for _, tr := range t.trackers {
		go tr.startAnnounceLoop(announce, firstEvent)
	}

	responseReceived := make(chan trackerAwareAnnounceResult, len(t.trackers))

	for _, tr := range t.trackers {
		go func(ti *AllTrackersTierAnnouncer, tr ITrackerAnnouncer) {
			for {
				select {
				case resp := <-tr.Responses():
					responseReceived <- resp
				case doneStopping := <-ti.stoppingLoops:
					tr.stopAnnounceLoop()
					doneStopping <- struct{}{}
					return
				}
			}
		}(t, tr)
	}

	trackersStates := make(map[uuid.UUID]tierState, len(t.trackers))
	// All trackers starts alive
	for _, tr := range t.trackers {
		trackersStates[tr.Uuid()] = ALIVE
	}
	currentTierState := ALIVE

	// this chan will be allocated once every time the tier changes his state to prevent spamming the receiver with the same message at every turn of fhe for loop.
	var stateUpdated chan tierState = nil

	for {
		select {
		case resp := <-responseReceived:
			var ts tierState = DEAD
			if resp.Err == nil {
				ts = ALIVE
			}
			trackersStates[resp.trackerUuid] = ts
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

			// if the state has changed make the chan non-nil to allow read from it (on the next loop it will go trough the case in the select
			stateUpdated = t.state
		case stateUpdated <- currentTierState:
			stateUpdated = nil
		case doneStopping := <-t.stoppingTier:
			doneStopping <- struct{}{}
			return
		}
	}
}

func (t *AllTrackersTierAnnouncer) stopAnnounceLoop() {
	wg := sync.WaitGroup{}

	for range t.trackers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			done := make(chan struct{})
			t.stoppingLoops <- done
			<-done
		}()
	}

	wg.Wait()

	// Stop the tier last. If a tracker routine write to the 'responseReceived' chan and the tier dont read it the routine will block and never goes into the stop case
	done := make(chan struct{})
	t.stoppingTier <- done
	<-done
}

type tierState = byte

const (
	ALIVE tierState = iota
	DEAD            = 0x01
)
