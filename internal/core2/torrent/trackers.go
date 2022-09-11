package torrent

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

type trackerPool struct {
	trackers                    []tracker
	announceToAllTiers          bool
	announceToAllTrackersInTier bool
}

func createTrackers(announce url.URL, announceList [][]url.URL, policy AnnouncePolicy) (trackerPool, error) {
	// TODO: the client should provide the announceList shuffled (depending on client behaviour)
	var trackers []tracker

	// add all announceList trackers (starting at tier 1)
	for tierIdx := range announceList {
		for trackerIdx := range announceList[tierIdx] {
			u := announceList[tierIdx][trackerIdx]

			trackers = append(trackers, tracker{
				url:              u,
				tier:             tierIdx + 1,
				nextAnnounce:     time.Now(),
				announcesHistory: []announceHistory{},
			})
		}
	}

	// prepend single announce tracker at tier 0
	if !policy.SupportAnnounceList() || len(trackers) == 0 {
		// disable all tier added from announceList
		for i := range trackers {
			trackers[i].disabled = announceListNotSupported
		}

		singleTracker := tracker{
			url:              announce,
			tier:             0,
			nextAnnounce:     time.Now(),
			announcesHistory: []announceHistory{},
		}

		// filter all occurrence of this new tracker from the announceList
		var filteredCopy []tracker
		for _, tr := range trackers {
			if tr.url.String() != singleTracker.url.String() {
				filteredCopy = append(filteredCopy, tr)
			}
		}

		trackers = append([]tracker{singleTracker}, filteredCopy...)
	}

	// disable trackers according to policy
	for i := range trackers {
		// may have been disabled because of announceList not supported
		if trackers[i].disabled.isDisabled() {
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

	return trackerPool{
		trackers:                    trackers,
		announceToAllTiers:          policy.ShouldAnnounceToAllTier(),
		announceToAllTrackersInTier: policy.ShouldAnnounceToAllTrackersInTier(),
	}, nil
}

//	 FIXME: ready to announce seems to be the only caller of findTrackersInUse, it might be possible to merge the two methods to prevent double array parsing + double array assignements
//		 Also, it may be possible to modify the signature of findInuse to trackerList findInuse([]tracker, announceToAllTiers bool, announceToAllTrackersInTier bool, filter func(tracker) bool), the filter being: if canAnnounce return true.
//		 One think to keep in mind though. If a tracker is "currentlyAnnouncing" it should be considered inUse, but not eligible to readyToAnnounce
func (ts *trackerPool) readyToAnnounce(at time.Time) []tracker {
	inUse := findTrackersInUse(ts.trackers, ts.announceToAllTiers, ts.announceToAllTrackersInTier)

	var ready []tracker
	for _, tr := range inUse {
		if tr.canAnnounce(at) {
			ready = append(ready, tr)
		}
	}

	return ready
}

func (ts *trackerPool) succeed(trackerUrl url.URL, response TrackerAnnounceResponse) {
	trackerIndex, err := findTrackerForUrl(ts.trackers, trackerUrl)
	if err != nil {
		return
	}

	ts.trackers[trackerIndex].announceSucceed(announceHistory{
		at:       time.Now(),
		interval: response.Interval,
		seeders:  response.Seeders,
		leechers: response.Leechers,
		error:    "",
	})
}

func (ts *trackerPool) failed(trackerUrl url.URL, response TrackerAnnounceResponseError) {
	trackerIndex, err := findTrackerForUrl(ts.trackers, trackerUrl)
	if err != nil {
		return
	}

	// operate on tracker before it gets moved in the list
	ts.trackers[trackerIndex].announceFailed(announceHistory{
		at:    time.Now(),
		error: response.Error.Error(),
	})

	deprioritizeTracker(ts.trackers, trackerIndex)
}

func findTrackerForUrl(trackerList []tracker, u url.URL) (int, error) {
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
func findTrackersInUse(trackerList []tracker, announceToAllTiers bool, announceToAllTrackersInTier bool) []tracker {
	var inUse []tracker

	// index of the tier we last found and inUse tracker in
	foundForTier := -1
	foundOne := false

	for i, tr := range trackerList {
		if tr.disabled.isDisabled() {
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
func deprioritizeTracker(trackers []tracker, indexToDeprioritize int) {
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
