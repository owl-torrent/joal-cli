package sharing

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSwarm_AddPeersToSwarm(t *testing.T) {
	swarm := NewSwarm()
	swarm.add("d", Peers{})

	assert.Len(t, swarm.peers, 1)
}

func TestSwarm_shouldReplacePeersByPeersIdentifier(t *testing.T) {
	const peerIdentifier = "my-key"
	swarm := Swarm{peers: map[string]Peers{
		peerIdentifier: Peers{seed: 0},
	}}

	newPeers := Peers{seed: 10}
	swarm.add(peerIdentifier, newPeers)

	assert.Equal(t, newPeers, swarm.peers[peerIdentifier])
}

func TestSwarm_shouldElectPeers(t *testing.T) {
	swarm := NewSwarm()
	swarm.add("1", Peers{leech: 10, seed: 5})
	swarm.add("2", Peers{leech: 15, seed: 6})
	swarm.add("3", Peers{leech: 20, seed: 10})

	assert.Equal(t, Peers{20, 10}, swarm.electMostRepresentativePeer(mostLeeched))
}

func TestSwarm_shouldElectZeroPeersWhenEmpty(t *testing.T) {
	assert.Equal(t, Peers{}, NewSwarm().electMostRepresentativePeer(mostLeeched))
}

func TestPeerElectionAlgorithm_mostLeeched_shouldElectMostLeeched(t *testing.T) {
	toBeElected := Peers{leech: 10}
	assert.Equal(t, toBeElected, mostLeeched(Peers{leech: 0}, toBeElected))
	assert.Equal(t, toBeElected, mostLeeched(toBeElected, Peers{leech: 0}))
}

func TestPeerElectionAlgorithm_MostLeechedNonZeroSeeders_shouldElectMostLeechedIfOnlyZeroSeeders(t *testing.T) {
	toBeElected := Peers{leech: 10}
	assert.Equal(t, toBeElected, MostLeechedNonZeroSeeders(Peers{leech: 0}, toBeElected))
	assert.Equal(t, toBeElected, MostLeechedNonZeroSeeders(toBeElected, Peers{leech: 0}))
}

func TestTestPeerElectionAlgorithm_MostLeechedNonZeroSeeders_shouldIgnoreZeroSeeders(t *testing.T) {
	toBeElected := Peers{leech: 10, seed: 1}
	assert.Equal(t, toBeElected, MostLeechedNonZeroSeeders(Peers{leech: 100}, toBeElected))
	assert.Equal(t, toBeElected, MostLeechedNonZeroSeeders(toBeElected, Peers{leech: 100}))
}

func TestPeerElectionAlgorithm_mostSeeded_shouldElectMostSeeded(t *testing.T) {
	toBeElected := Peers{seed: 10}
	assert.Equal(t, toBeElected, mostSeeded(Peers{seed: 0}, toBeElected))
	assert.Equal(t, toBeElected, mostSeeded(toBeElected, Peers{seed: 0}))
}

func TestPeerElectionAlgorithm_MostSeededNonZeroLeechers_shouldElectMostSeededIfOnlyZeroLeechers(t *testing.T) {
	toBeElected := Peers{seed: 10}
	assert.Equal(t, toBeElected, MostSeededNonZeroLeechers(Peers{seed: 0}, toBeElected))
	assert.Equal(t, toBeElected, MostSeededNonZeroLeechers(toBeElected, Peers{seed: 0}))
}

func TestTestPeerElectionAlgorithm_MostSeededNonZeroLeechers_shouldIgnoreZeroLeechers(t *testing.T) {
	toBeElected := Peers{leech: 1, seed: 10}
	assert.Equal(t, toBeElected, MostSeededNonZeroLeechers(Peers{seed: 100}, toBeElected))
	assert.Equal(t, toBeElected, MostSeededNonZeroLeechers(toBeElected, Peers{seed: 100}))
}

/*
func TestSwarm_ShouldElectMostRepresentative(t *testing.T) {
	swarm := Swarm{peers: map[string]Peers{
		"1": Peers{seed: 0, leech: 0},
		"2": Peers{seed: 0, leech: 0},
		"3": Peers{seed: 0, leech: 0},
		"4": Peers{seed: 0, leech: 0},
	}}

	swarm.elect(mostSeeded)

}*/
