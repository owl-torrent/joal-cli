package sharing

type Swarm struct {
	peers map[string]Peers
}

func NewSwarm() *Swarm {
	return &Swarm{peers: map[string]Peers{}}
}

func (s *Swarm) add(peerIdentifier string, peers Peers) {
	s.peers[peerIdentifier] = peers
}

func (s *Swarm) electMostRepresentativePeer(election PeerElectionAlgorithm) Peers {
	elected := Peers{}

	for _, peers := range s.peers {
		elected = election(elected, peers)
	}

	return elected
}

type Peers struct {
	leech int
	seed  int
}

func (p Peers) Leechers() int {
	return p.leech
}

func (p Peers) Seeders() int {
	return p.seed
}

type PeerElectionAlgorithm func(candidate1 Peers, candidate2 Peers) Peers

// mostLeeched return the torrent with the most leechers.
var mostLeeched PeerElectionAlgorithm = func(candidate1 Peers, candidate2 Peers) Peers {
	if candidate1.leech > candidate2.leech {
		return candidate1
	}
	return candidate2
}

// MostLeechedNonZeroSeeders return the torrent with the most leechers. Peers with 0 seeders are not considered valid candidates
var MostLeechedNonZeroSeeders PeerElectionAlgorithm = func(candidate1 Peers, candidate2 Peers) Peers {
	if candidate1.seed == candidate2.seed {
		return mostLeeched(candidate1, candidate2)
	}
	if candidate1.seed == 0 && candidate2.seed > 0 {
		return candidate2
	}
	if candidate2.seed == 0 && candidate1.seed > 0 {
		return candidate1
	}

	return mostLeeched(candidate1, candidate2)
}

// mostSeeded return the torrent with the most seeders.
var mostSeeded PeerElectionAlgorithm = func(candidate1 Peers, candidate2 Peers) Peers {
	if candidate1.seed > candidate2.seed {
		return candidate1
	}
	return candidate2
}

// MostSeededNonZeroLeechers return the torrent with the most seeders. Peers with 0 leechers are not considered valid candidates
var MostSeededNonZeroLeechers PeerElectionAlgorithm = func(candidate1 Peers, candidate2 Peers) Peers {
	if candidate1.leech == candidate2.leech {
		return mostSeeded(candidate1, candidate2)
	}
	if candidate1.leech == 0 && candidate2.leech > 0 {
		return candidate2
	}
	if candidate2.leech == 0 && candidate1.leech > 0 {
		return candidate1
	}

	return mostSeeded(candidate1, candidate2)
}
