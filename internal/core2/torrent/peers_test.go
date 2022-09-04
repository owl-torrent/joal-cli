package torrent

import (
	"github.com/anthonyraymond/joal-cli/internal/utils/testutils"
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

func TestPeersElectorGetPeersZeroByDefault(t *testing.T) {
	elector := newPeersElector()

	peers := elector.GetPeers()
	assert.Equal(t, int32(0), peers.Seeders)
	assert.Equal(t, int32(0), peers.Leechers)
}

func TestPeersElector_UpdatePeersForTrackerShouldAddEntry(t *testing.T) {
	elector := newPeersElector()

	elector.UpdatePeersForTracker(PeersUpdateRequest{
		trackerUrl: *testutils.MustParseUrl("http://localhost:1/announce"),
		Seeders:    10,
		Leechers:   20,
	})
	elector.UpdatePeersForTracker(PeersUpdateRequest{
		trackerUrl: *testutils.MustParseUrl("http://localhost:2/announce"),
		Seeders:    10,
		Leechers:   20,
	})

	assert.Len(t, elector.allPeers, 2)
}

func TestPeersElector_UpdatePeersForTrackerShouldElectAfterAdding(t *testing.T) {
	elector := newPeersElector()
	elector.allPeers[peersIdentifierFromUrl(*testutils.MustParseUrl("http://localhost:1/announce"))] = Peers{}

	elector.UpdatePeersForTracker(PeersUpdateRequest{
		trackerUrl: *testutils.MustParseUrl("http://localhost:1/announce"),
		Seeders:    10,
		Leechers:   20,
	})

	elected := elector.GetPeers()
	assert.EqualValues(t, 10, elected.Seeders)
	assert.EqualValues(t, 20, elected.Leechers)
}

func TestPeersElector_UpdatePeersForTrackerShouldReplacingEntry(t *testing.T) {
	elector := newPeersElector(mostLeeched)

	elector.UpdatePeersForTracker(PeersUpdateRequest{
		trackerUrl: *testutils.MustParseUrl("http://localhost:1/announce"),
		Seeders:    10,
		Leechers:   20,
	})

	elector.UpdatePeersForTracker(PeersUpdateRequest{
		trackerUrl: *testutils.MustParseUrl("http://localhost:1/announce"),
		Seeders:    0,
		Leechers:   10,
	})

	assert.Len(t, elector.allPeers, 1)
}

func TestPeersElector_RemovePeersForTracker(t *testing.T) {
	elector := newPeersElector()

	elector.allPeers[peersIdentifierFromUrl(*testutils.MustParseUrl("http://localhost:1/announce"))] = Peers{}
	assert.Len(t, elector.allPeers, 1)

	elector.RemovePeersForTracker(PeersDeleteRequest{
		trackerUrl: *testutils.MustParseUrl("http://localhost:1/announce"),
	})

	assert.Len(t, elector.allPeers, 0)
}

func TestPeersElector_RemovePeersForTrackerShouldElectAfterRemoving(t *testing.T) {
	elector := newPeersElector()
	elector.allPeers[peersIdentifierFromUrl(*testutils.MustParseUrl("http://localhost:1/announce"))] = Peers{}
	elector.allPeers[peersIdentifierFromUrl(*testutils.MustParseUrl("http://localhost:9000/announce"))] = Peers{
		Seeders:  50,
		Leechers: 50,
	}

	// ensure the election has not been processed yet
	assert.EqualValues(t, 0, elector.GetPeers().Leechers)

	elector.RemovePeersForTracker(PeersDeleteRequest{
		trackerUrl: *testutils.MustParseUrl("http://localhost:1/announce"),
	})

	// ensure the election has not been processed yet
	elected := elector.GetPeers()
	assert.EqualValues(t, 50, elected.Seeders)
	assert.EqualValues(t, 50, elected.Leechers)
}

func Test_electionAlgorithm_electMostLeechedShouldPreferTheBiggestAmountOfLeechers(t *testing.T) {
	candidate1 := Peers{Seeders: 10, Leechers: 100}
	candidate2 := Peers{Seeders: 1000, Leechers: 50}

	assert.EqualValues(t, candidate1, mostLeeched(candidate1, candidate2))
	assert.EqualValues(t, candidate1, mostLeeched(candidate2, candidate1))
}

func Test_electionAlgorithm_electMostLeechedShouldPreferTheBiggestAmountOfLeechersWithAtLeastOneSeeders(t *testing.T) {
	candidate1 := Peers{Seeders: 10, Leechers: 100}
	candidate2 := Peers{Seeders: 0, Leechers: 10000000}

	assert.EqualValues(t, candidate1, mostLeeched(candidate1, candidate2))
	assert.EqualValues(t, candidate1, mostLeeched(candidate2, candidate1))
}

func Test_electionAlgorithm_electMostLeechedShouldPreferTheBiggestAmountOfLeechersIfAllHaveZeroSeeders(t *testing.T) {
	candidate1 := Peers{Seeders: 0, Leechers: 100}
	candidate2 := Peers{Seeders: 0, Leechers: 50}

	assert.EqualValues(t, candidate1, mostLeeched(candidate1, candidate2))
	assert.EqualValues(t, candidate1, mostLeeched(candidate2, candidate1))
}

func Test_electPeersShouldElectEmptyArray(t *testing.T) {
	expectedWinner := Peers{Seeders: 0, Leechers: 0}
	var candidates map[string]Peers

	var noopAlgo electionAlgorithm = func(p1 Peers, p2 Peers) Peers {
		return p1
	}

	elected := electPeers(candidates, noopAlgo)
	assert.EqualValues(t, expectedWinner, elected)
}

func Test_peersIdentifierFromUrl(t *testing.T) {
	type args struct {
		u url.URL
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "with port",
			args: args{u: *testutils.MustParseUrl("http://localhost:/announce")},
			want: "http://localhost:",
		},
		{
			name: "without port",
			args: args{u: *testutils.MustParseUrl("http://localhost:8080/announce")},
			want: "http://localhost:8080",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, peersIdentifierFromUrl(tt.args.u), "peersIdentifierFromUrl(%v)", tt.args.u)
		})
	}
}
