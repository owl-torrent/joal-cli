package torrent

import (
	"errors"
	"github.com/anthonyraymond/joal-cli/internal/utils/testutils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTorrent_GetPeers(t *testing.T) {
	tor := torrent{peers: newPeersElector(mostLeeched)}

	assert.EqualValues(t, 0, tor.GetPeers().Seeders())
	assert.EqualValues(t, 0, tor.GetPeers().Leechers())

	tor.peers.electedPeer = peers{
		seeders:  10,
		leechers: 20,
	}
	assert.EqualValues(t, 10, tor.GetPeers().Seeders())
	assert.EqualValues(t, 20, tor.GetPeers().Leechers())
}

func TestTorrent_HandleAnnounceSuccess(t *testing.T) {
	tor := torrent{
		contrib: contribution{},
		peers:   newPeersElector(mostLeeched),
		trackers: trackerPool{
			trackers: []tracker{},
		},
	}

	tor.HandleAnnounceSuccess(TrackerAnnounceResponse{
		Request:  TrackerAnnounceRequest{Url: *testutils.MustParseUrl("http://localhost:8081")},
		Interval: 1,
		Seeders:  2,
		Leechers: 3,
	})

	// should update peers
	assert.EqualValues(t, 2, tor.GetPeers().Seeders())
	assert.EqualValues(t, 3, tor.GetPeers().Leechers())
}

func TestTorrent_HandleAnnounceFailed(t *testing.T) {
	tor := torrent{
		contrib: contribution{},
		peers:   newPeersElector(mostLeeched),
		trackers: trackerPool{
			trackers: []tracker{},
		},
	}

	// add some peers for this tracker
	tor.peers.updatePeersForTracker(peersUpdateRequest{
		trackerUrl: *testutils.MustParseUrl("http://localhost:8081"),
		Seeders:    40,
		Leechers:   50,
	})
	assert.EqualValues(t, 40, tor.GetPeers().Seeders())
	assert.EqualValues(t, 50, tor.GetPeers().Leechers())

	tor.HandleAnnounceError(TrackerAnnounceResponseError{
		Error:   errors.New("announce has failed"),
		Request: TrackerAnnounceRequest{Url: *testutils.MustParseUrl("http://localhost:8081")},
	})

	// should have removed peers
	assert.EqualValues(t, 0, tor.GetPeers().Seeders())
	assert.EqualValues(t, 0, tor.GetPeers().Leechers())
}
