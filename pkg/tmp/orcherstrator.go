package tmp

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"math/rand"
	"net/url"
	"time"
)

type TierAnnouncer interface {
}

type AllTiersAnnouncer struct {
	tiers []TrackersAnnouncer
}

type FallbackTiersAnnouncer struct {
	tiers []TrackersAnnouncer
}

type TrackersAnnouncer interface {
}

type AllTrackersAnnouncer struct {
	trackers []url.URL
}

type FallbackTrackersAnnouncer struct {
	trackers []url.URL
}

type AnnounceOrchestrator struct {
	t             *Torrent
	tierAnnouncer TierAnnouncer
}

func NewAnnounceOrchestrator(t *Torrent, announceToAllTiers bool, announceToAllTrackersInTier bool) (*AnnounceOrchestrator, error) {
	var annList [][]string = t.metainfo.AnnounceList

	if t.metainfo.AnnounceList != nil && len(t.metainfo.AnnounceList) == 0 {
		// If torrent does not have an announce-list, use the old & announce property
		annList = [][]string{{t.metainfo.Announce}}
	}

	tierAnnouncer, err := newTierAnnouncer(annList, announceToAllTiers, announceToAllTrackersInTier)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create announce orchestrator for torrent '%s'", t.info.Name)
	}

	return &AnnounceOrchestrator{
		t:             t,
		tierAnnouncer: tierAnnouncer,
	}, nil
}

func newTierAnnouncer(announceList [][]string, announceToAllTiers bool, announceToAllTrackersInTier bool) (TierAnnouncer, error) {
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
			var tier TierAnnouncer
			if announceToAllTrackersInTier {
				tier = &AllTrackersAnnouncer{
					trackers: trackers,
				}
			} else {
				tier = &FallbackTrackersAnnouncer{
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
			tiers: tiers,
		}, nil
	} else {
		return &FallbackTiersAnnouncer{
			tiers: tiers,
		}, nil
	}
}

func (a AnnounceOrchestrator) Run() {

}

func Stop() {

}
