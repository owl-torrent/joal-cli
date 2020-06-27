package torrent

import (
	"github.com/anacrolix/torrent/tracker"
	"github.com/google/uuid"
	"sync"
)

type AllTrackersTierAnnouncer struct {
	uuid         uuid.UUID
	trackers     []ITrackerAnnouncer
	state        chan tierState
	stoppingLoop chan chan struct{}
}

func newAllTrackersTierAnnouncer(trackers ...ITrackerAnnouncer) ITierAnnouncer {
	t := &AllTrackersTierAnnouncer{
		uuid:         uuid.New(),
		trackers:     trackers,
		state:        make(chan tierState),
		stoppingLoop: make(chan chan struct{}),
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

	lock := &sync.RWMutex{}
	trackersStates := make(map[uuid.UUID]tierState)
	// All trackers starts alive
	for _, tr := range t.trackers {
		trackersStates[tr.Uuid()] = ALIVE
	}

	for _, tr := range t.trackers {
		go func(ti *AllTrackersTierAnnouncer, tr ITrackerAnnouncer) {
			for {
				select {
				case resp := <-tr.Responses():
					if resp.Err != nil {
						lock.Lock()
						trackersStates[resp.trackerUuid] = DEAD
						lock.Unlock()
						break
					}
					lock.Lock()
					trackersStates[resp.trackerUuid] = ALIVE
					lock.Unlock()
				case doneStopping := <-ti.stoppingLoop:
					tr.stopAnnounceLoop()
					doneStopping <- struct{}{}
					return
				}

				// if any tracker is alive, consider the tier alive
				var state tierState = DEAD
				lock.RLock()
				for _, tState := range trackersStates {
					if tState == ALIVE {
						state = ALIVE
						break
					}
				}
				lock.RUnlock()

				select {
				case ti.state <- state:
				case doneStopping := <-ti.stoppingLoop:
					tr.stopAnnounceLoop()
					doneStopping <- struct{}{}
					return
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
