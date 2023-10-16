package torrent

import (
	"fmt"
	"net/url"
)

var (
	mostLeeched electionAlgorithm = func(candidate1 peers, candidate2 peers) peers {
		// The two below conditions ensure that the returned peer won't have 0 seeders if one of the two candidate has 1 seeders
		if candidate1.seeders == 0 && candidate2.seeders > 0 {
			return candidate2
		}
		if candidate2.seeders == 0 && candidate1.seeders > 0 {
			return candidate1
		}

		if candidate1.leechers < candidate2.leechers {
			return candidate2
		}
		return candidate1
	}
)

type electionAlgorithm = func(p1 peers, p2 peers) peers

type peersElector struct {
	allPeers    map[string]peers
	electedPeer peers
	algorithm   electionAlgorithm
}

func newPeersElector(algorithm ...electionAlgorithm) peersElector {
	selectedAlgorithm := mostLeeched
	if len(algorithm) != 0 {
		selectedAlgorithm = algorithm[0]
	}

	return peersElector{
		allPeers:    make(map[string]peers),
		electedPeer: peers{},
		algorithm:   selectedAlgorithm,
	}
}

func (e *peersElector) GetPeers() Peers {
	return e.electedPeer
}

func (e *peersElector) updatePeersForTracker(r peersUpdateRequest) peers {
	e.allPeers[peersIdentifierFromUrl(r.trackerUrl)] = peers{
		seeders:  r.Seeders,
		leechers: r.Leechers,
	}

	e.electedPeer = electPeers(e.allPeers, e.algorithm)

	return e.electedPeer
}

func (e *peersElector) removePeersForTracker(r peersDeleteRequest) peers {
	delete(e.allPeers, peersIdentifierFromUrl(r.trackerUrl))

	e.electedPeer = electPeers(e.allPeers, e.algorithm)

	return e.electedPeer
}

func peersIdentifierFromUrl(u url.URL) string {
	return fmt.Sprintf("%s://%s", u.Scheme, u.Host)
}

func electPeers(candidates map[string]peers, elect electionAlgorithm) peers {
	var elected peers
	if len(candidates) == 0 {
		return elected
	}

	for _, candidate := range candidates {
		// if currently elected is zero value assign and continue
		if elected.leechers == 0 && elected.seeders == 0 {
			elected = candidate
			continue
		}
		elected = elect(elected, candidate)
	}
	return elected
}

type peersUpdateRequest struct {
	trackerUrl url.URL
	Seeders    int32
	Leechers   int32
}

type peersDeleteRequest struct {
	trackerUrl url.URL
}

type peers struct {
	seeders  int32
	leechers int32
}

func (p peers) Seeders() int32 {
	return p.seeders
}

func (p peers) Leechers() int32 {
	return p.leechers
}
