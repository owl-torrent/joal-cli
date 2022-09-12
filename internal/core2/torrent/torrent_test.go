package torrent

import (
	"errors"
	"github.com/anacrolix/torrent/metainfo"
	libtracker "github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/internal/utils/testutils"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTorrent_InfoHash(t *testing.T) {
	hash := metainfo.NewHashFromHex("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")
	tor := torrent{infoHash: hash}

	assert.Equal(t, hash, tor.InfoHash())
}

func TestTorrent_Name(t *testing.T) {
	name := "my name"
	tor := torrent{info: slimInfo{name: name}}

	assert.Equal(t, name, tor.Name())
}

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

func TestTorrent_HandleAnnounceError(t *testing.T) {
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

func TestTorrent_AnnounceToReadyTrackers(t *testing.T) {
	tor := torrent{
		contrib: contribution{
			uploaded:   10,
			downloaded: 20,
			left:       30,
			corrupt:    40,
		},
		peers: newPeersElector(mostLeeched),
		trackers: trackerPool{
			announceToAllTrackersInTier: true,
			announceToAllTiers:          true,
			trackers: []tracker{
				{url: *testutils.MustParseUrl("http://localhost:8081"), nextAnnounce: time.Now().Add(-1 * time.Minute)},
				{url: *testutils.MustParseUrl("http://localhost:8082"), nextAnnounce: time.Now().Add(1 * time.Hour)},
			},
		},
	}

	var calledForUrl []string

	announcingFunc := func(request TrackerAnnounceRequest) {
		calledForUrl = append(calledForUrl, request.Url.String())
		assert.EqualValues(t, 10, request.Uploaded)
		assert.EqualValues(t, 20, request.Downloaded)
		assert.EqualValues(t, 30, request.Left)
		assert.EqualValues(t, 40, request.Corrupt)
	}

	tor.AnnounceToReadyTrackers(announcingFunc)

	// announce only to tracker available
	assert.Len(t, calledForUrl, 1)
	assert.Equal(t, "http://localhost:8081", calledForUrl[0])
}

func TestTorrent_AnnounceStop(t *testing.T) {
	tor := torrent{
		contrib: contribution{
			uploaded:   10,
			downloaded: 20,
			left:       30,
			corrupt:    40,
		},
		peers: newPeersElector(mostLeeched),
		trackers: trackerPool{
			announceToAllTrackersInTier: true,
			announceToAllTiers:          true,
			trackers: []tracker{
				{url: *testutils.MustParseUrl("http://localhost:8081"), startSent: true, nextAnnounce: time.Now().Add(-1 * time.Minute)},
				{url: *testutils.MustParseUrl("http://localhost:8082"), startSent: true, nextAnnounce: time.Now().Add(1 * time.Hour)},
			},
		},
	}

	var calledForUrl []string

	announcingFunc := func(request TrackerAnnounceRequest) {
		calledForUrl = append(calledForUrl, request.Url.String())
		assert.EqualValues(t, libtracker.Stopped, request.Event)
		assert.EqualValues(t, 10, request.Uploaded)
		assert.EqualValues(t, 20, request.Downloaded)
		assert.EqualValues(t, 30, request.Left)
		assert.EqualValues(t, 40, request.Corrupt)
	}

	tor.AnnounceStop(announcingFunc)

	assert.Len(t, calledForUrl, 2)
	assert.Contains(t, calledForUrl, "http://localhost:8081")
	assert.Contains(t, calledForUrl, "http://localhost:8082")
}

func TestTorrent_withRequestEnhancer(t *testing.T) {
	tor := torrent{
		infoHash: metainfo.NewHashFromHex("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"),
		info:     slimInfo{private: true},
		contrib: contribution{
			uploaded:   10,
			downloaded: 20,
			left:       30,
			corrupt:    40,
		},
	}

	called := false
	announcingFunc := func(request TrackerAnnounceRequest) {
		called = true
		assert.EqualValues(t, tor.infoHash.Bytes(), request.InfoHash.Bytes())
		assert.EqualValues(t, 10, request.Uploaded)
		assert.EqualValues(t, 20, request.Downloaded)
		assert.EqualValues(t, 30, request.Left)
		assert.EqualValues(t, 40, request.Corrupt)
		assert.True(t, request.Private)
	}

	withRequestEnhancer(announcingFunc, tor)(TrackerAnnounceRequest{})
	assert.True(t, called)
}
