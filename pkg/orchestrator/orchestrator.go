package orchestrator

//go:generate mockgen -destination=./orchestrator_mock.go -self_package=github.com/anthonyraymond/joal-cli/pkg/orchestrator -package=orchestrator github.com/anthonyraymond/joal-cli/pkg/orchestrator ITrackerAnnouncer,ITierAnnouncer

import (
	"context"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/announcer"
	"github.com/anthonyraymond/joal-cli/pkg/logs"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/url"
	"strings"
	"sync"
	"time"
)

type AnnouncingFunction = func(ctx context.Context, u url.URL, event tracker.AnnounceEvent) (announcer.AnnounceResponse, error)
type tierState = byte
type trackerState = tierState

const (
	ALIVE tierState = iota
	DEAD            = 0x01
)

type trackerAnnounceResult struct {
	Err       error
	Interval  time.Duration
	Completed time.Time
}

type ITrackerAnnouncer interface {
	announceOnce(ctx context.Context, announce AnnouncingFunction, event tracker.AnnounceEvent) trackerAnnounceResult
	startAnnounceLoop(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent)
	Responses() <-chan trackerAnnounceResult
	stopAnnounceLoop()
}

type ITierAnnouncer interface {
	announceOnce(ctx context.Context, announce AnnouncingFunction, event tracker.AnnounceEvent) tierState
	startAnnounceLoop(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent)
	States() <-chan tierState
	LastKnownInterval() (time.Duration, error)
	stopAnnounceLoop()
}

type IOrchestrator interface {
	Start(announce AnnouncingFunction)
	Stop(context context.Context, announce AnnouncingFunction)
}

type FallbackOrchestrator struct {
	tier           *linkedTierList
	stopping       chan chan struct{}
	loopInProgress bool
	lock           *sync.RWMutex
}

type IConfig interface {
	DoesSupportAnnounceList() bool
	ShouldAnnounceToAllTiers() bool
	ShouldAnnounceToAllTrackersInTier() bool
}

func NewOrchestrator(meta metainfo.MetaInfo, conf IConfig) (IOrchestrator, error) {
	log := logs.GetLogger()
	if conf == nil {
		return nil, errors.New("nil orchestrator config")
	}

	if !conf.DoesSupportAnnounceList() {
		log.Info("build orchestrator without support for announce-list", zap.String("url", meta.Announce))
		var announceList = [][]string{{meta.Announce}}
		return createOrchestratorForAnnounceList(announceList, true, true)
	}

	if !meta.AnnounceList.OverridesAnnounce(meta.Announce) {
		log.Info("build orchestrator with 'announce' because 'announce-list' is empty", zap.String("url", meta.Announce))
		var announceList = [][]string{{meta.Announce}}
		return createOrchestratorForAnnounceList(announceList, true, true)
	}

	// dont trust your inputs: some url (or even tiers) may be empty, filter them
	var announceList [][]string
	for _, tier := range meta.AnnounceList {
		tiers := make([]string, 0)
		for _, u := range tier {
			if strings.TrimSpace(u) != "" {
				tier = append(tier, u)
			}
		}
		if len(tiers) > 0 {
			announceList = append(announceList, tiers)
		}
	}

	if len(announceList) == 0 {
		return nil, errors.New("announce-list is empty")
	}

	if !conf.DoesSupportAnnounceList() {
		log.Info("build orchestrator without support for announce-list", zap.String("url", meta.Announce))
		var announceList = [][]string{{meta.Announce}}
		return createOrchestratorForAnnounceList(announceList, true, true)
	}
	log.Info("build orchestrator with 'announce-list'", zap.Any("announce-list", announceList))
	return createOrchestratorForAnnounceList(announceList, conf.ShouldAnnounceToAllTiers(), conf.ShouldAnnounceToAllTrackersInTier())
}

func createOrchestratorForAnnounceList(announceList [][]string, announceToAllTiers bool, announceToAllTrackersInTier bool) (IOrchestrator, error) {
	var tiers []ITierAnnouncer

	for _, tier := range announceList {
		var trackers []ITrackerAnnouncer
		for _, trackerUrl := range tier {
			u, err := url.Parse(trackerUrl)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to parse torrent Announce uri '%s'", trackerUrl)
			}
			t := newTracker(*u)
			trackers = append(trackers, t)
		}

		var tier ITierAnnouncer
		var err error
		if announceToAllTrackersInTier {
			tier, err = newAllTrackersTierAnnouncer(trackers...)
		} else {
			tier, err = newFallbackTrackersTierAnnouncer(trackers...)
		}
		if err != nil {
			return nil, errors.Wrap(err, "failed to create tier")
		}
		tiers = append(tiers, tier)
	}

	var o IOrchestrator
	var err error
	if announceToAllTiers {
		o, err = newAllOrchestrator(tiers...)
	} else {
		o, err = newFallBackOrchestrator(tiers...)
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to create orchestrator")
	}

	return o, nil
}

func newFallBackOrchestrator(tiers ...ITierAnnouncer) (IOrchestrator, error) {
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

func (o *FallbackOrchestrator) Stop(ctx context.Context, annFunc AnnouncingFunction) {
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
			tier.announceOnce(ctx, annFunc, tracker.Stopped)
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

func newAllOrchestrator(tiers ...ITierAnnouncer) (IOrchestrator, error) {
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

func (o *AllOrchestrator) Stop(ctx context.Context, annFunc AnnouncingFunction) {
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
				tier.announceOnce(ctx, annFunc, tracker.Stopped)
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
