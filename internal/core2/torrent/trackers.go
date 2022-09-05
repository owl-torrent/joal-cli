package torrent

import (
	"fmt"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anthonyraymond/joal-cli/internal/core2/common"
	"math/rand"
	"net/url"
	"strings"
	"time"
)

var (
	ErrAllTrackerAreDisabled = fmt.Errorf("trackers.all-trackers-disabled")
)

type Trackers struct {
	trackers                    []Tracker
	announceToAllTiers          bool
	announceToAllTrackersInTier bool
}

// FIXME: this implementation may be more simple if we had url.URL instead of string and [][]string in constructor, see if caller can Parse URL instead

// TODO: the client should provide the announceList shuffled? that way he can decide to shuffle or not (depending on client configuration)
func CreateTrackers(announce string, announceList metainfo.AnnounceList, policy AnnouncePolicy) (Trackers, error) {
	var trackers []Tracker

	announceList = createShuffledCopyOfAnnounceList(announceList)
	// Shuffling trackers of announceList according to BEP-12: https://www.bittorrent.org/beps/bep_0012.html
	rand.Seed(time.Now().UnixNano())
	for _, tier := range announceList {
		rand.Shuffle(len(tier), func(i, j int) {
			tier[i], tier[j] = tier[j], tier[i]
		})
	}

	// add all announceList trackers (starting at tier 1)
	for tierIdx := range announceList {
		for trackerIdx := range announceList[tierIdx] {
			u, err := url.Parse(announceList[tierIdx][trackerIdx])
			if err != nil {
				//FIXME: log failed to parse tracker
				continue
			}

			trackers = append(trackers, Tracker{
				url:              *u,
				tier:             tierIdx + 1,
				nextAnnounce:     time.Now(),
				announcesHistory: []AnnounceHistory{},
			})
		}
	}

	// prepend single announce tracker at tier 0
	if !policy.SupportAnnounceList() || len(trackers) == 0 {
		// disable all tier added from announceList
		for i := range trackers {
			trackers[i].disabled = announceListNotSupported
		}

		u, err := url.Parse(announce)
		if err == nil {
			singleTracker := Tracker{
				url:              *u,
				tier:             0,
				nextAnnounce:     time.Now(),
				announcesHistory: []AnnounceHistory{},
			}

			// filter all occurrence of this new tracker from the announceList
			var filteredCopy []Tracker
			for _, tr := range trackers {
				if tr.url.String() != singleTracker.url.String() {
					filteredCopy = append(filteredCopy, tr)
				}
			}

			trackers = append([]Tracker{singleTracker}, filteredCopy...)
		} else {
			//FIXME: log failed to parse URL
		}
	}

	// disable trackers acording to policy
	for i := range trackers {
		// may have been disabled because of announceList not supported
		if trackers[i].disabled.IsDisabled() {
			continue
		}
		// disable udp according to policy
		if !policy.SupportUdpAnnounce() && strings.HasPrefix(trackers[i].url.Scheme, "udp") {
			trackers[i].disabled = announceProtocolNotSupported
		}
		// disable http according to policy
		if !policy.SupportHttpAnnounce() && strings.HasPrefix(trackers[i].url.Scheme, "http") {
			trackers[i].disabled = announceProtocolNotSupported
		}
		// disable any other funky protocols
		if !strings.HasPrefix(trackers[i].url.Scheme, "http") && !strings.HasPrefix(trackers[i].url.Scheme, "udp") {
			trackers[i].disabled = announceProtocolNotSupported
		}
	}

	if !hasOneEnabled(trackers) {
		return Trackers{}, ErrAllTrackerAreDisabled
	}

	return Trackers{
		trackers:                    trackers,
		announceToAllTiers:          policy.ShouldAnnounceToAllTier(),
		announceToAllTrackersInTier: policy.ShouldAnnounceToAllTrackersInTier(),
	}, nil
}

// Return all tracker that may be used given announce policy
func findTrackersInUse(trackerList []Tracker, announceToAllTiers bool, announceToAllTrackersInTier bool) []Tracker {
	panic(fmt.Errorf("not implemented"))
}

func (ts *Trackers) ReadyToAnnounce(time.Time) []Tracker {
	// return inUse & ready to announce
	panic(fmt.Errorf("not implemented"))
}

func (ts *Trackers) Succeed(trackerUrl url.URL, response common.AnnounceResponse) {
	panic(fmt.Errorf("not implemented"))
}

func (ts *Trackers) Failed(trackerUrl url.URL, response common.AnnounceResponseError) {
	// deprioritizeTracker()
	panic(fmt.Errorf("not implemented"))
}

func findTrackerForUrl(trackerList []Tracker, u url.URL) (index int, t Tracker, err error) {
	panic(fmt.Errorf("not implemented"))
}

func deprioritizeTracker(trackers []Tracker, indexToDeprioritize int) []Tracker {
	panic(fmt.Errorf("not implemented"))
}

func hasOneEnabled(trackers []Tracker) bool {
	for _, tracker := range trackers {
		if !tracker.disabled.IsDisabled() {
			return true
		}
	}

	return false
}

func createShuffledCopyOfAnnounceList(list [][]string) [][]string {
	duplicate := make([][]string, len(list))
	for i := range list {
		duplicate[i] = make([]string, len(list[i]))
		copy(duplicate[i], list[i])
	}
	return duplicate
}
