package orchestrator

import "sync"

type tierState = byte

const (
	DEAD  tierState = 0
	ALIVE tierState = 1
)

// A TierStateCalculator is a wrapper around multiple providers and send events when the global state of all his children has changed.
// if all
type ITierStateCalculator interface {
	StateChan() <-chan tierState
}

type tierStateCalculator struct {
	firstStateReported bool
	globalState        tierState
	trackersLastState  map[ITrackerAnnouncer]bool
	globalStateChan    chan tierState
	lock               *sync.Mutex
}

func NewTierStateCalculator(stateProviders []ITrackerAnnouncer) *tierStateCalculator {
	c := &tierStateCalculator{
		firstStateReported: false,
		globalState:        ALIVE,
		trackersLastState:  map[ITrackerAnnouncer]bool{},
		globalStateChan:    make(chan tierState, 1),
		lock:               &sync.Mutex{},
	}

	for _, provider := range stateProviders {
		c.trackersLastState[provider] = true
	}

	return c
}

func (c *tierStateCalculator) setIndividualState(provider ITrackerAnnouncer, alive bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.trackersLastState[provider] = alive

	if alive && !c.firstStateReported {
		c.firstStateReported = true
		c.globalStateChan <- c.globalState
		return
	}

	// If the new value is equal to the global state there is no need to check if the new value will change the state
	if (alive && c.globalState == ALIVE) || (!alive && c.globalState == DEAD) {
		return
	}

	// recalculate the tier state with the new information available
	var stateAfterUpdate tierState = DEAD
	for _, trackerAlive := range c.trackersLastState {
		if trackerAlive {
			stateAfterUpdate = ALIVE
			break
		}
	}
	// unchanged
	if stateAfterUpdate == c.globalState {
		return
	}

	c.globalState = stateAfterUpdate
	c.globalStateChan <- c.globalState
	c.firstStateReported = true
}

func (c *tierStateCalculator) States() <-chan tierState {
	return c.globalStateChan
}

func (c *tierStateCalculator) Destroy() {
	c.lock.Lock()
	defer c.lock.Unlock()

	drainStateChannel(c.globalStateChan)
	close(c.globalStateChan)
}

func drainStateChannel(c <-chan tierState) {
	for {
		select {
		case <-c:
		default:
			return
		}
	}
}
