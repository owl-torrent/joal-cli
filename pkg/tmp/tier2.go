package tmp

import (
	"github.com/anacrolix/torrent/tracker"
	"github.com/google/uuid"
	"sync"
)

type AllTrackersTierAnnouncer struct {
	uuid           uuid.UUID
	trackers       []ITrackerAnnouncer
	trackersStates map[uuid.UUID]tierState
	state          chan tierState
	stoppingLoop   chan chan struct{}
}

func newAllTrackersTierAnnouncer(trackers []ITrackerAnnouncer) ITierAnnouncer {
	t := &AllTrackersTierAnnouncer{
		uuid:           uuid.New(),
		trackers:       trackers,
		trackersStates: nil,
		state:          make(chan tierState),
		stoppingLoop:   make(chan chan struct{}),
	}

	// All trackers starts alive
	t.trackersStates = make(map[uuid.UUID]tierState)
	for _, tr := range trackers {
		t.trackersStates[tr.Uuid()] = ALIVE
	}

	return t
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

	for _, tr := range t.trackers {
		go func(t *AllTrackersTierAnnouncer, tr ITrackerAnnouncer) {
			for {
				select {
				case resp := <-tr.Responses():
					if resp.Err != nil {
						t.trackersStates[resp.trackerUuid] = DEAD
						break
					}
					t.trackersStates[resp.trackerUuid] = ALIVE
				case doneStopping := <-t.stoppingLoop:
					t.stopAnnounceLoop()
					doneStopping <- struct{}{}
				}

				// if any tracker is alive, consider the tier alive
				var state tierState = DEAD
				for _, tState := range t.trackersStates {
					if tState == ALIVE {
						state = ALIVE
						break
					}
				}

				select {
				case t.state <- state:
				case doneStopping := <-t.stoppingLoop:
					t.stopAnnounceLoop()
					doneStopping <- struct{}{}
				}
			}
		}(t, tr)
	}
}

func (t *AllTrackersTierAnnouncer) stopAnnounceLoop() {
	wg := sync.WaitGroup{}

	for range t.trackers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			done := make(chan struct{})
			t.stoppingLoop <- done
			<-done
		}()
	}

	wg.Wait()
}

type tierState = byte

const (
	ALIVE tierState = iota
	DEAD            = 0x01
)
