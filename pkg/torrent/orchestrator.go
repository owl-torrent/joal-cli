package torrent

//go:generate mockgen -destination=./orchestrator_mock.go -self_package=github.com/anthonyraymond/joal-cli/pkg/torrent -package=torrent github.com/anthonyraymond/joal-cli/pkg/torrent Orchestrator,ITrackerAnnouncer,ITierAnnouncer

import (
	"context"
	"errors"
	"github.com/anacrolix/torrent/tracker"
	"github.com/google/uuid"
	"net/url"
	"time"
)

type AnnouncingFunction = func(u url.URL, event tracker.AnnounceEvent, ctx context.Context) trackerAnnounceResult

type ITrackerAnnouncer interface {
	Uuid() uuid.UUID
	announceOnce(announce AnnouncingFunction, event tracker.AnnounceEvent) trackerAwareAnnounceResult
	startAnnounceLoop(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent)
	Responses() <-chan trackerAwareAnnounceResult
	stopAnnounceLoop()
}

type ITierAnnouncer interface {
	Uuid() uuid.UUID
	announceOnce(announce AnnouncingFunction, event tracker.AnnounceEvent) tierState
	startAnnounceLoop(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent)
	States() <-chan tierState
	LastKnownInterval() (time.Duration, error)
	stopAnnounceLoop()
}

type Orchestrator interface {
	Start(announce AnnouncingFunction)
	Stop(context context.Context)
}

type FallbackBackOrchestrator struct {
	tier     *linkedTierList
	stopping chan chan struct{}
}

func NewFallBackOrchestrator(tiers ...ITierAnnouncer) (Orchestrator, error) {
	if len(tiers) == 0 {
		return nil, errors.New("tiers list can not be empty")
	}
	list, err := newLinkedTierList(tiers)
	if err != nil {
		return nil, err
	}
	return &FallbackBackOrchestrator{
		tier:     list,
		stopping: make(chan chan struct{}),
	}, nil
}

func (o *FallbackBackOrchestrator) Start(announce AnnouncingFunction) {
	go func(o *FallbackBackOrchestrator) {
		startAnnounceTiers := time.After(0 * time.Millisecond)

		currentEvent := tracker.Started

		for {
			select {
			case <-startAnnounceTiers:
				startAnnounceTiers = nil
				event := currentEvent
				go o.tier.startAnnounceLoop(announce, event)
			case st := <-o.tier.States():
				if st == DEAD {
					o.tier.stopAnnounceLoop()
					drainStatesChannel(o.tier) // ensure no more event are queued. Otherwise next time we use next and get back to this tier we might have an old message
					o.tier.next()
					startAnnounceTiers = time.After(0 * time.Millisecond)
					if o.tier.isFirst() { // we have travel through the whole list and get back to the first tier, lets wait before trying to re-announce on the first tier
						interval, err := o.tier.LastKnownInterval()
						if err != nil {
							interval = DefaultDurationWaitOnError
						}
						startAnnounceTiers = time.After(interval)
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
					startAnnounceTiers = time.After(interval)
					o.tier.rewindToFirst()
				}
			case doneStopping := <-o.stopping:
				o.tier.stopAnnounceLoop()
				drainStatesChannel(o.tier)
				doneStopping <- struct{}{}
				return
			}
		}
	}(o)
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

func (o *FallbackBackOrchestrator) Stop(ctx context.Context) {
	done := make(chan struct{})
	o.stopping <- done
	<-done

	// TODO: announceStop once (and fallback to next till is succeed, but return if all fails once)
}

type tierState = byte
type trackerState = tierState

const (
	ALIVE tierState = iota
	DEAD            = 0x01
)
