package tmp

import (
	"github.com/anacrolix/torrent/tracker"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"math"
	"math/rand"
	"net/url"
	"sync"
	"time"
)

type TiersAnnouncer interface {
	startAnnouncing()
	awaitTermination()
}

type TrackersAnnouncer interface {
	startAnnouncing()
	awaitTermination()
}

// TODO: il ne faudrait pas que les TrackersAnnouncer créer des goroutines, car le TiersAnnouncer a besoin de feedback pour orchestrer les autres tiers (backup, ...) Il faudrait que ca fasse les annonces et que ca renvoi le resultat. Puis on fera une pause arbitraire d'une interval renvoyé par un des tracker.

type AnnounceOrchestrator struct {
	t              *Torrent
	tiersAnnouncer TiersAnnouncer
}
type AnnouncingFunction = func(u url.URL, event tracker.AnnounceEvent) trackerAnnounceResult

func NewAnnounceOrchestrator(t *Torrent, announceToAllTiers bool, announceToAllTrackersInTier bool) (*AnnounceOrchestrator, error) {
	var annList [][]string = t.metaInfo.AnnounceList

	if t.metaInfo.AnnounceList != nil && len(t.metaInfo.AnnounceList) == 0 {
		// If torrent does not have an announce-list, use the old & announce property
		annList = [][]string{{t.metaInfo.Announce}}
	}

	tierAnnouncer, err := newTierAnnouncer(t, annList, announceToAllTiers, announceToAllTrackersInTier)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create announce orchestrator for torrent '%s'", t.info.Name)
	}

	return &AnnounceOrchestrator{
		t:              t,
		tiersAnnouncer: tierAnnouncer,
	}, nil
}

func newTierAnnouncer(t *Torrent, announceList [][]string, announceToAllTiers bool, announceToAllTrackersInTier bool) (TiersAnnouncer, error) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	var tiers []TrackersAnnouncer
	for _, tier := range announceList {
		var trackers []url.URL

		shuffledTierUrls := append([]string(nil), tier...)
		// Shuffle the tiers URI once at the beginning according to: https://www.bittorrent.org/beps/bep_0012.html
		rng.Shuffle(len(shuffledTierUrls), func(i, j int) { shuffledTierUrls[i], shuffledTierUrls[j] = shuffledTierUrls[j], shuffledTierUrls[i] })

		for _, trackerUrl := range shuffledTierUrls {
			u, err := url.Parse(trackerUrl)
			if err != nil {
				logrus.WithField("url", trackerUrl).WithError(err).Warn("Failed to parse tracker url")
				continue
			}
			trackers = append(trackers, *u)
		}

		if len(trackers) > 0 { // if the tier has no URL do not add to tier list
			var tier TiersAnnouncer
			if announceToAllTrackersInTier {
				tier = &AllTrackersAnnouncer{
					shutdownWg: &sync.WaitGroup{},
					t:          t,
					trackers:   trackers,
				}
			} else {
				tier = &FallbackTrackersAnnouncer{
					shutdownWg: &sync.WaitGroup{},
					t:          t,
					trackers:   trackers,
				}
			}
			tiers = append(tiers, tier)
		}
	}

	if len(tiers) == 0 {
		return nil, errors.New("the torrent does not have any valid trackers")
	}

	if announceToAllTiers {
		return &AllTiersAnnouncer{
			t:     t,
			tiers: tiers,
		}, nil
	} else {
		return &FallbackTiersAnnouncer{
			t:     t,
			tiers: tiers,
		}, nil
	}
}

func (a *AnnounceOrchestrator) Run() {

}

func (a *AnnounceOrchestrator) AwaitTermination() {
	a.tiersAnnouncer.awaitTermination()
}

type AllTrackersAnnouncer struct {
	shutdownWg *sync.WaitGroup
	t          *Torrent
	trackers   []url.URL
}

func (a *AllTrackersAnnouncer) startAnnouncing() {
	for _, u := range a.trackers {
		go func(u url.URL) {
			a.shutdownWg.Add(1)
			defer a.shutdownWg.Done()
			defer func() {
				_ = a.t.announce(u, tracker.Stopped)
			}()

			// create a mocked last announce with a default interval
			lastAnnounce := trackerAnnounceResult{Interval: 5 * time.Minute, Completed: time.Now()}
			event := tracker.Started

			announceResult := a.t.announce(u, event)
			if announceResult.Err != nil {
				announceResult.Interval = calculateNextAnnounceDelayAfterError(lastAnnounce)
			} else {
				// after first successfully announce, get back to regular "none"
				event = tracker.None
			}

			lastAnnounce = announceResult

			select {
			case <-a.t.closed.C(): // When the torrents closes, also close the orchestrator
				return
			case <-time.After(time.Until(announceResult.Completed.Add(announceResult.Interval))):
			}
		}(u)
	}
}

func (a *AllTrackersAnnouncer) awaitTermination() {
	a.shutdownWg.Wait()
}

type FallbackTrackersAnnouncer struct {
	shutdownWg *sync.WaitGroup
	t          *Torrent
	trackers   []url.URL
}

func (a *FallbackTrackersAnnouncer) startAnnouncing() {
}

func (a *FallbackTrackersAnnouncer) awaitTermination() {
	a.shutdownWg.Wait()
}

type AllTiersAnnouncer struct {
	t        *Torrent
	tiers    []TrackersAnnouncer
	announce AnnouncingFunction
}

func (a *AllTiersAnnouncer) startAnnouncing() {

}

func (a *AllTiersAnnouncer) awaitTermination() {
	for _, tier := range a.tiers {
		tier.awaitTermination()
	}
}

type FallbackTiersAnnouncer struct {
	t        *Torrent
	tiers    []TrackersAnnouncer
	announce AnnouncingFunction
}

func (a *FallbackTiersAnnouncer) startAnnouncing() {
}

func (a *FallbackTiersAnnouncer) awaitTermination() {
	for _, tier := range a.tiers {
		tier.awaitTermination()
	}
}

func calculateNextAnnounceDelayAfterError(lastAnnounce trackerAnnounceResult) time.Duration {
	// On first error, retry fast
	if lastAnnounce.Err == nil {
		return 5 * time.Minute
	} else { // Otherwise, double the delay up to a maximum of 1800s
		delay := math.Min(1800, (lastAnnounce.Interval * 2).Seconds())
		return time.Duration(delay) * time.Second
	}
}
