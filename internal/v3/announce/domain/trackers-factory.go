package domain

import (
	"fmt"
	"github.com/anacrolix/torrent/metainfo"
	"math/rand"
	"net/url"
	"strings"
	"time"
)

var randSeed = time.Now().UnixNano()

type AnnounceAbility struct {
	SupportHttpAnnounce bool
	SupportUdpAnnounce  bool
	SupportAnnounceList bool
}

func CreateTrackers(announce string, announceList metainfo.AnnounceList, ability AnnounceAbility) (Trackers, error) {
	announceUrl, err := announceUriOf(announce)
	if err != nil {
		return Trackers{}, err
	}

	announceListUrls, err := announceListOf(
		shuffle(announceList),
	)
	if err != nil {
		return Trackers{}, err
	}

	// Build the trackers with all the announceList urls

	var trackers []Tracker

	for tierIdx := range announceListUrls {
		for trackerIdx := range announceListUrls[tierIdx] {
			u := announceListUrls[tierIdx][trackerIdx]

			trackers = append(trackers, Tracker{
				Url:     u,
				Tier:    tierIdx + 1, // add 1 to start the tier count at 1 instead of 0
				History: []AnnounceHistory{},
			})
		}
	}

	// if the AnnounceAbility does not support AnnounceList, we prepend our tiers with a tier "0" containing the single Announce url
	if !ability.SupportAnnounceList || len(trackers) == 0 {
		// disable all tier added from announceList
		for i := range trackers {
			trackers[i].State.Disable = announceListNotSupported
		}

		singleTracker := Tracker{
			Url:     announceUrl,
			Tier:    0,
			History: []AnnounceHistory{},
		}

		// filter all occurrence of this new tracker from the announceList
		var filteredCopy []Tracker
		for _, tr := range trackers {
			if tr.Url.String() != singleTracker.Url.String() {
				filteredCopy = append(filteredCopy, tr)
			}
		}

		trackers = append([]Tracker{singleTracker}, filteredCopy...)
	}

	// disable trackers according to AnnounceAbility
	for i := range trackers {
		// may have been disabled because of announceList not supported
		if trackers[i].State.isDisabled() {
			continue
		}
		// disable udp according to policy
		if !ability.SupportUdpAnnounce && strings.HasPrefix(trackers[i].Url.Scheme, "udp") {
			trackers[i].State.Disable = announceProtocolNotSupported
		}
		// disable http according to policy
		if !ability.SupportHttpAnnounce && strings.HasPrefix(trackers[i].Url.Scheme, "http") {
			trackers[i].State.Disable = announceProtocolNotSupported
		}
		// disable any other funky protocols
		if !strings.HasPrefix(trackers[i].Url.Scheme, "http") && !strings.HasPrefix(trackers[i].Url.Scheme, "udp") {
			trackers[i].State.Disable = announceProtocolNotSupported
		}
	}

	return Trackers{
		trackers: trackers,
	}, nil
}

func announceUriOf(announceUri string) (url.URL, error) {
	u, err := url.Parse(announceUri)
	if err != nil {
		return url.URL{}, fmt.Errorf("failed to parse announce url property from torrent metadata: %w", err)
	}
	return *u, nil
}

func announceListOf(announceList [][]string) ([][]url.URL, error) {
	result := make([][]url.URL, len(announceList))

	for tier := range announceList {
		result[tier] = make([]url.URL, len(announceList[tier]))

		for track := range announceList[tier] {
			u, err := url.Parse(announceList[tier][track])
			if err != nil {
				return nil, fmt.Errorf("failed to parse url in announceList '%s': %w", announceList[tier][track], err)
			}
			result[tier][track] = *u
		}
	}

	return result, nil
}

// Shuffle an announce list according to BEP-12: https://www.bittorrent.org/beps/bep_0012.html
func shuffle(announceList metainfo.AnnounceList) metainfo.AnnounceList {
	copiedAnnounceList := announceList.Clone()

	rand.Seed(randSeed)
	for _, tier := range copiedAnnounceList {
		rand.Shuffle(len(tier), func(i, j int) {
			tier[i], tier[j] = tier[j], tier[i]
		})
	}
	return copiedAnnounceList
}
