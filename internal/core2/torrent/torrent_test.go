package torrent

import (
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
