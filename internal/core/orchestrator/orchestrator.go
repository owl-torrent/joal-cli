package orchestrator

import (
	"context"
	"fmt"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/internal/core/announcer"
	"github.com/anthonyraymond/joal-cli/internal/core/logs"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/url"
	"strings"
	"sync"
	"time"
)

type AnnouncingFunction = func(ctx context.Context, u url.URL, event tracker.AnnounceEvent) (announcer.AnnounceResponse, error)

type trackerAnnounceResult struct {
	Err       error
	Interval  time.Duration
	Completed time.Time
}

type ITrackerAnnouncer interface {
	announceOnce(ctx context.Context, announce AnnouncingFunction, event tracker.AnnounceEvent) trackerAnnounceResult
	startAnnounceLoop(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan trackerAnnounceResult, error)
	stopAnnounceLoop()
}

type ITierAnnouncer interface {
	announceOnce(ctx context.Context, announce AnnouncingFunction, event tracker.AnnounceEvent) tierState
	startAnnounceLoop(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) (<-chan tierState, error)
	LastKnownInterval() time.Duration
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

type TorrentInfo struct {
	Announce     string
	AnnounceList metainfo.AnnounceList
}

func NewOrchestrator(info *TorrentInfo, conf IConfig) (IOrchestrator, error) {
	log := logs.GetLogger()
	if conf == nil {
		return nil, fmt.Errorf("nil orchestrator config")
	}

	if !conf.DoesSupportAnnounceList() {
		log.Info("build orchestrator without support for announce-list", zap.String("url", info.Announce))
		return createOrchestratorForAnnounceList([][]string{{info.Announce}}, true, true)
	}

	// AnnounceList contains only empty url but announce contains a valid url
	if !info.AnnounceList.OverridesAnnounce(info.Announce) {
		log.Info("build orchestrator with 'announce' because 'announce-list' is empty", zap.String("url", info.Announce))
		return createOrchestratorForAnnounceList([][]string{{info.Announce}}, true, true)
	}

	// dont trust your inputs: some url (or even tiers) may be empty, filter them
	var announceList [][]string
	for _, tier := range info.AnnounceList {
		currentTier := make([]string, 0)
		for _, trackerUri := range tier {
			if strings.TrimSpace(trackerUri) != "" {
				currentTier = append(currentTier, trackerUri)
			}
		}
		if len(currentTier) > 0 {
			announceList = append(announceList, currentTier)
		}
	}

	if len(announceList) == 0 {
		return nil, fmt.Errorf("announce-list is empty")
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
		return nil, fmt.Errorf("tiers list can not be empty")
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
	defer o.lock.Unlock()
	if o.loopInProgress {
		return
	}
	o.loopInProgress = true

	go func() {
		pauseBeforeLoop := time.After(0 * time.Millisecond)
		currentEvent := tracker.Started
		var tierStates <-chan tierState = nil

		for {
			select {
			case <-pauseBeforeLoop:
				pauseBeforeLoop = nil
				event := currentEvent
				var err error
				tierStates, err = o.tier.startAnnounceLoop(announce, event)
				if err != nil {
					o.tier.stopAnnounceLoop()
					tierStates = nil
					o.tier.next()
					pauseBeforeLoop = time.After(5 * time.Second)
				}
			case st := <-tierStates:
				if st == DEAD {
					o.tier.stopAnnounceLoop()
					tierStates = nil
					o.tier.next()
					pauseBeforeLoop = time.After(0 * time.Millisecond)
					if o.tier.isFirst() { // we have travel through the whole list and get back to the first tier, lets wait before trying to re-announce on the first tier
						pauseBeforeLoop = time.After(o.tier.LastKnownInterval())
					}
					break
				}
				currentEvent = tracker.None // as soon an event succeed we can proceed with None all the subsequent time

				if !o.tier.isFirst() { // A backup tier has successfully announced, lets get back to primary tier
					o.tier.stopAnnounceLoop()
					tierStates = nil
					pauseBeforeLoop = time.After(o.tier.LastKnownInterval())
					o.tier.backToFirst()
				}
			case doneStopping := <-o.stopping:
				o.tier.stopAnnounceLoop()
				doneStopping <- struct{}{}
				return
			}
		}
	}()
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

	stopAnnounceCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	waitChan := make(chan struct{})
	go func(tier ITierAnnouncer) {
		defer close(waitChan)
		tier.announceOnce(stopAnnounceCtx, annFunc, tracker.Stopped)
	}(o.tier)

	<-waitChan
}

type AllOrchestrator struct {
	tiers          []ITierAnnouncer
	stopping       chan chan struct{}
	loopInProgress bool
	lock           *sync.RWMutex
}

func newAllOrchestrator(tiers ...ITierAnnouncer) (IOrchestrator, error) {
	if len(tiers) == 0 {
		return nil, fmt.Errorf("tiers list can not be empty")
	}
	return &AllOrchestrator{
		tiers:          tiers,
		stopping:       make(chan chan struct{}),
		loopInProgress: false,
		lock:           &sync.RWMutex{},
	}, nil
}

func (o *AllOrchestrator) Start(announce AnnouncingFunction) {
	o.lock.Lock()
	defer o.lock.Unlock()
	if o.loopInProgress {
		return
	}
	o.loopInProgress = true

	stoppingLoops := make(chan chan struct{}, len(o.tiers))

	for _, t := range o.tiers {
		go func(t ITierAnnouncer) {
			tierStates, err := t.startAnnounceLoop(announce, tracker.Started)
			if err != nil {
				t.stopAnnounceLoop()
				logs.GetLogger().Warn("failed to start announcing", zap.Error(err))
			}

			for {
				select {
				case <-tierStates:
					// dont care about the answer. All the tiers need to keeps announcing no matter what. We just want to consume the channel to prevent deadlock
				case doneStopping := <-stoppingLoops:
					t.stopAnnounceLoop()
					doneStopping <- struct{}{}
					return
				}
			}
		}(t)
	}

	go func() {
		doneStopping := <-o.stopping
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
	}()
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

	stopAnnounceCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	waitChan := make(chan struct{})
	go func() {
		wg := &sync.WaitGroup{}
		for _, tier := range o.tiers {
			wg.Add(1)
			go func(tier ITierAnnouncer) {
				defer wg.Done()
				tier.announceOnce(stopAnnounceCtx, annFunc, tracker.Stopped)
			}(tier)
		}
		wg.Wait()
		close(waitChan)
	}()

	<-waitChan
}
