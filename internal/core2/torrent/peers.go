package torrent

import (
	"fmt"
	"net/url"
)

var (
	mostLeeched electionAlgorithm = func(candidate1 Peers, candidate2 Peers) Peers {
		// The two below conditions ensure that the returned peer won't have 0 seeders if one of the two candidate has 1 seeders
		if candidate1.Seeders == 0 && candidate2.Seeders > 0 {
			return candidate2
		}
		if candidate2.Seeders == 0 && candidate1.Seeders > 0 {
			return candidate1
		}

		if candidate1.Leechers < candidate2.Leechers {
			return candidate2
		}
		return candidate1
	}
)

type electionAlgorithm = func(p1 Peers, p2 Peers) Peers

type peersElector struct {
	allPeers    map[string]Peers
	electedPeer Peers
	algorithm   electionAlgorithm
}

func newPeersElector(algorithm ...electionAlgorithm) peersElector {
	selectedAlgorithm := mostLeeched
	if len(algorithm) != 0 {
		selectedAlgorithm = algorithm[0]
	}

	return peersElector{
		allPeers:    make(map[string]Peers),
		electedPeer: Peers{},
		algorithm:   selectedAlgorithm,
	}
}

func (e *peersElector) GetPeers() Peers {
	return e.electedPeer
}

func (e *peersElector) UpdatePeersForTracker(r PeersUpdateRequest) Peers {
	e.allPeers[peersIdentifierFromUrl(r.trackerUrl)] = Peers{
		Seeders:  r.Seeders,
		Leechers: r.Leechers,
	}

	e.electedPeer = electPeers(e.allPeers, e.algorithm)

	return e.electedPeer
}

func (e *peersElector) RemovePeersForTracker(r PeersDeleteRequest) Peers {
	delete(e.allPeers, peersIdentifierFromUrl(r.trackerUrl))

	e.electedPeer = electPeers(e.allPeers, e.algorithm)

	return e.electedPeer
}

func peersIdentifierFromUrl(u url.URL) string {
	return fmt.Sprintf("%s://%s", u.Scheme, u.Host)
}

func electPeers(candidates map[string]Peers, elect electionAlgorithm) Peers {
	var elected Peers
	if len(candidates) == 0 {
		return elected
	}

	for _, candidate := range candidates {
		// if currently elected is zero value assign and continue
		if elected.Leechers == 0 && elected.Seeders == 0 {
			elected = candidate
			continue
		}
		elected = elect(elected, candidate)
	}
	return elected
}

type PeersUpdateRequest struct {
	trackerUrl url.URL
	Seeders    int32
	Leechers   int32
}

type PeersDeleteRequest struct {
	trackerUrl url.URL
}

type Peers struct {
	Seeders  int32
	Leechers int32
}
