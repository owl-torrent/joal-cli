package tmp

import (
	"github.com/anacrolix/torrent/tracker"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"math"
	"math/rand"
	"net/url"
	"time"
)

type TiersAnnouncer interface {
}

type AllTiersAnnouncer struct {
	t        *Torrent
	tiers    []TrackersAnnouncer
	announce TorrentAwareAnnounceFunc
}

type FallbackTiersAnnouncer struct {
	t        *Torrent
	tiers    []TrackersAnnouncer
	announce TorrentAwareAnnounceFunc
}

type TrackersAnnouncer interface {
}

type AllTrackersAnnouncer struct {
	t        *Torrent
	trackers []url.URL
	announce TorrentAwareAnnounceFunc
}

type FallbackTrackersAnnouncer struct {
	t        *Torrent
	trackers []url.URL
	announce TorrentAwareAnnounceFunc
}

type AnnounceOrchestrator struct {
	t              *Torrent
	tiersAnnouncer TiersAnnouncer
}
type TorrentAwareAnnounceFunc = func(u url.URL, event tracker.AnnounceEvent) trackerAnnounceResult

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
					t:        t,
					trackers: trackers,
				}
			} else {
				tier = &FallbackTrackersAnnouncer{
					t:        t,
					trackers: trackers,
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

func (a AnnounceOrchestrator) Run() {

}

func (a AnnounceOrchestrator) Stop() {

}

func (a *AllTrackersAnnouncer) startAnnouncing(announce TorrentAwareAnnounceFunc) {
	for _, u := range a.trackers {
		go func(announce TorrentAwareAnnounceFunc, u url.URL) {
			defer func() { _ = announce(u, tracker.Stopped) }() //TODO: this may cause a problem, it is executed after the torrent has closed his chan, and some resources may already have been released, i can result in a panic at some point

			// create a mocked last announce with a default interval
			lastAnnounce := trackerAnnounceResult{Interval: 5 * time.Minute, Completed: time.Now()}
			event := tracker.Started

			announceResult := announce(u, event)
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
		}(announce, u)
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
