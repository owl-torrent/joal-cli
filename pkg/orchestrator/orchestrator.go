package orchestrator

//go:generate mockgen -destination=./orchestrator_mock.go -self_package=github.com/anthonyraymond/joal-cli/pkg/orchestrator -package=orchestrator github.com/anthonyraymond/joal-cli/pkg/orchestrator Orchestrator,ITrackerAnnouncer,ITierAnnouncer

import (
	"context"
	"errors"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/seed"
	"net/url"
	"sync"
	"time"
)

type AnnouncingFunction = func(u url.URL, event tracker.AnnounceEvent, ctx context.Context) seed.TrackerAnnounceResult
type tierState = byte
type trackerState = tierState

const (
	ALIVE tierState = iota
	DEAD            = 0x01
)

type ITrackerAnnouncer interface {
	announceOnce(announce AnnouncingFunction, event tracker.AnnounceEvent) seed.TrackerAnnounceResult
	startAnnounceLoop(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent)
	Responses() <-chan seed.TrackerAnnounceResult
	stopAnnounceLoop()
}

type ITierAnnouncer interface {
	announceOnce(announce AnnouncingFunction, event tracker.AnnounceEvent) tierState
	startAnnounceLoop(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent)
	States() <-chan tierState
	LastKnownInterval() (time.Duration, error)
	stopAnnounceLoop()
}

type Orchestrator interface {
	Start(announce AnnouncingFunction)
	Stop(announce AnnouncingFunction, context context.Context)
}

type FallbackOrchestrator struct {
	tier           *linkedTierList
	stopping       chan chan struct{}
	loopInProgress bool
	lock           *sync.RWMutex
}

func NewFallBackOrchestrator(tiers ...ITierAnnouncer) (Orchestrator, error) {
	if len(tiers) == 0 {
		return nil, errors.New("tiers list can not be empty")
	}
	list, err := newLinkedTierList(tiers)
	if err != nil {
		return nil, err
	}
	return &FallbackOrchestrator{
		tier:           list,
		stopping:       make(chan chan struct{}),
		loopInProgress: false,
		lock:           &sync.RWMutex{},
	}, nil
}

func (o *FallbackOrchestrator) Start(announce AnnouncingFunction) {
	o.lock.Lock()
	if o.loopInProgress {
		o.lock.Unlock()
		return
	}
	o.loopInProgress = true
	o.lock.Unlock()

	pauseBeforeLoop := time.After(0 * time.Millisecond)

	currentEvent := tracker.Started

	for {
		select {
		case <-pauseBeforeLoop:
			pauseBeforeLoop = nil
			event := currentEvent
			go o.tier.startAnnounceLoop(announce, event)
		case st := <-o.tier.States():
			if st == DEAD {
				o.tier.stopAnnounceLoop()
				drainStatesChannel(o.tier) // ensure no more event are queued. Otherwise next time we use next and get back to this tier we might have an old message
				o.tier.next()
				pauseBeforeLoop = time.After(0 * time.Millisecond)
				if o.tier.isFirst() { // we have travel through the whole list and get back to the first tier, lets wait before trying to re-announce on the first tier
					interval, err := o.tier.LastKnownInterval()
					if err != nil {
						interval = DefaultDurationWaitOnError
					}
					pauseBeforeLoop = time.After(interval)
				}
				break
			}
			currentEvent = tracker.None // as soon an event succeed we can proceed with None all the subsequent time

			if !o.tier.isFirst() { // A backup tier has successfully announced, lets get back to primary tier
				o.tier.stopAnnounceLoop()
				drainStatesChannel(o.tier) // ensure no more event are queued. Otherwise next time we use next and get back to this tier we might have an old message
				interval, err := o.tier.LastKnownInterval()
				if err != nil {
					interval = DefaultDurationWaitOnError
				}
				pauseBeforeLoop = time.After(interval)
				o.tier.backToFirst()
			}
		case doneStopping := <-o.stopping:
			o.tier.stopAnnounceLoop()
			drainStatesChannel(o.tier)
			doneStopping <- struct{}{}
			return
		}
	}
}

func drainStatesChannel(t ITierAnnouncer) {
	for {
		select {
		case <-t.States():
			continue
		default:
			return
		}
	}
}

func (o *FallbackOrchestrator) Stop(annFunc AnnouncingFunction, ctx context.Context) {
	o.lock.Lock()
	defer o.lock.Unlock()
	if !o.loopInProgress {
		return
	}
	o.loopInProgress = false

	done := make(chan struct{})
	o.stopping <- done

	select { // both case just going, if context is expired we still want to do the rest (which is non blocking and will return almost instantaneously
	case <-done:
	case <-ctx.Done():
	}

	waitChan := make(chan struct{})
	go func() {
		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func(tier ITierAnnouncer) {
			defer wg.Done()
			tier.announceOnce(annFunc, tracker.Stopped)
		}(o.tier)

		wg.Wait()
		close(waitChan)
	}()

	select { // both case just going, if context is expired we still want to do the rest (which is non blocking and will return almost instantaneously
	case <-waitChan:
	case <-ctx.Done():
		// TODO: log exit context done
	}
}

type AllOrchestrator struct {
	tiers          []ITierAnnouncer
	stopping       chan chan struct{}
	loopInProgress bool
	lock           *sync.RWMutex
}

func NewAllOrchestrator(tiers ...ITierAnnouncer) (Orchestrator, error) {
	if len(tiers) == 0 {
		return nil, errors.New("tiers list can not be empty")
	}
	return &AllOrchestrator{
		tiers:          tiers,
		stopping:       make(chan chan struct{}),
		loopInProgress: false,
		lock:           &sync.RWMutex{},
	}, nil
}

func (o *AllOrchestrator) Start(announce AnnouncingFunction) {
	type tierAwareState struct {
		state tierState
		tier  ITierAnnouncer
	}

	o.lock.Lock()
	if o.loopInProgress {
		o.lock.Unlock()
		return
	}
	o.loopInProgress = true
	o.lock.Unlock()

	for _, tr := range o.tiers {
		go tr.startAnnounceLoop(announce, tracker.Started)
	}

	stateReceived := make(chan tierAwareState, len(o.tiers))
	stoppingLoops := make(chan chan struct{}, len(o.tiers))

	for _, t := range o.tiers {
		go func(t ITierAnnouncer) {
			for {
				select {
				case resp := <-t.States():
					stateReceived <- tierAwareState{state: resp, tier: t}
				case doneStopping := <-stoppingLoops:
					t.stopAnnounceLoop()
					doneStopping <- struct{}{}
					return
				}
			}
		}(t)
	}

	for {
		select {
		case <-stateReceived:
			// dont give a **** about the answer. All the tiers need to keeps announcing no matter what. We just want to consume the chanel to prevent deadlock
		case doneStopping := <-o.stopping:
			wg := sync.WaitGroup{}

			for range o.tiers {
				wg.Add(1)
				go func() {
					defer wg.Done()
					done := make(chan struct{})
					stoppingLoops <- done
					<-done
				}()
			}
			wg.Wait()
			doneStopping <- struct{}{}

			return
		}
	}
}

func (o *AllOrchestrator) Stop(annFunc AnnouncingFunction, ctx context.Context) {
	o.lock.Lock()
	defer o.lock.Unlock()
	if !o.loopInProgress {
		return
	}
	o.loopInProgress = false

	done := make(chan struct{})
	o.stopping <- done

	select { // both case just going, if context is expired we still want to do the rest (which is non blocking and will return almost instantaneously
	case <-done:
	case <-ctx.Done():
	}

	waitChan := make(chan struct{})
	go func() {
		wg := &sync.WaitGroup{}
		for _, tier := range o.tiers {
			wg.Add(1)
			go func(tier ITierAnnouncer) {
				defer wg.Done()
				tier.announceOnce(annFunc, tracker.Stopped)
			}(tier)
		}

		wg.Wait()
		close(waitChan)
	}()

	select { // both case just going, if context is expired we still want to do the rest (which is non blocking and will return almost instantaneously
	case <-waitChan:
	case <-ctx.Done():
		// TODO: log exit context done
	}
}
