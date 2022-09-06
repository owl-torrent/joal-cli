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

	// disable trackers according to policy
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

//	 FIXME: ready to announce seems to be the only caller of findTrackersInUse, it might be possible to merge the two methods to prevent double array parsing + double rray assignements
//		 Also, it may be possible to modify the signature of findInuse to trackerList findInuse([]Tracker, announceToAllTiers bool, announceToAllTrackersInTier bool, filter func(Tracker) bool), the filter being: if canAnnounce return true.
//		 One think to keep in mind though. If a tracker is "currentlyAnnouncing" it should be considered inUse, but not eligible to readyToAnnounce
func (ts *Trackers) ReadyToAnnounce(at time.Time) []Tracker {
	inUse := findTrackersInUse(ts.trackers, ts.announceToAllTiers, ts.announceToAllTrackersInTier)

	var ready []Tracker
	for _, tracker := range inUse {
		if tracker.canAnnounce(at) {
			ready = append(ready, tracker)
		}
	}

	return ready
}

func (ts *Trackers) Succeed(trackerUrl url.URL, response common.AnnounceResponse) {
	trackerIndex, err := findTrackerForUrl(ts.trackers, trackerUrl)
	if err != nil {
		return
	}

	ts.trackers[trackerIndex].announceSucceed(AnnounceHistory{
		at:       time.Now(),
		interval: response.Interval,
		seeders:  response.Seeders,
		leechers: response.Leechers,
		error:    "",
	})
}

func (ts *Trackers) Failed(trackerUrl url.URL, response common.AnnounceResponseError) {
	trackerIndex, err := findTrackerForUrl(ts.trackers, trackerUrl)
	if err != nil {
		return
	}

	// operate on tracker before it gets moved in the list
	ts.trackers[trackerIndex].announceFailed(AnnounceHistory{
		at:    time.Now(),
		error: response.Error.Error(),
	})

	deprioritizeTracker(ts.trackers, trackerIndex)
}

func findTrackerForUrl(trackerList []Tracker, u url.URL) (int, error) {
	for i := range trackerList {
		if strings.EqualFold(trackerList[i].url.String(), u.String()) {
			return i, nil
		}
	}

	return -1, fmt.Errorf("no tracker of url '%s' in list", u.String())
}

// Return all tracker that may be used given announce policy
//   - AnnounceToAllTier = true , announceToAllTrackersInTier = true  => all (non-disabled) trackers
//   - AnnounceToAllTier = true , announceToAllTrackersInTier = false => first (non-disabled) tracker of each tier (if not first tracker it means that the tracker has been deprioritized)
//   - AnnounceToAllTier = false, announceToAllTrackersInTier = true => all (non-disabled) trackers of the first tier that contains a (non-disabled) tracker
//   - AnnounceToAllTier = false, announceToAllTrackersInTier = false => first (non-disabled) tracker
func findTrackersInUse(trackerList []Tracker, announceToAllTiers bool, announceToAllTrackersInTier bool) []Tracker {
	var inUse []Tracker

	// index of the tier we last found and inUse tracker in
	foundForTier := -1
	foundOne := false

	for i, tr := range trackerList {
		if tr.disabled.IsDisabled() {
			continue
		}
		if announceToAllTiers && !announceToAllTrackersInTier && foundForTier == tr.tier {
			continue
		}
		// Announcing to a single tracker in a single tier => we found one => exit
		if !announceToAllTiers && !announceToAllTrackersInTier && foundOne {
			return inUse
		}
		// Announcing to all trackers in one tier => we have found at least one and changed tier => exit
		if !announceToAllTiers && announceToAllTrackersInTier && foundOne && i > 0 && tr.tier > trackerList[i-1].tier {
			return inUse
		}

		foundOne = true
		foundForTier = tr.tier
		inUse = append(inUse, tr)
	}

	return inUse
}

// deprioritizeTracker push a tracker pointed by indexToDeprioritize to the end of his tier.
func deprioritizeTracker(trackers []Tracker, indexToDeprioritize int) {
	if indexToDeprioritize >= len(trackers)-1 {
		// out of bound or already the last one
		return
	}
	trackerToDeprioritize := trackers[indexToDeprioritize]

	for i := indexToDeprioritize; i < len(trackers); i++ {
		if i+1 == len(trackers) {
			return
		}
		t := trackers[i]
		if t.tier > trackerToDeprioritize.tier {
			return
		}

		if trackers[i+1].tier == trackerToDeprioritize.tier {
			// swap
			trackers[i], trackers[i+1] = trackers[i+1], trackers[i]
		}
	}
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
