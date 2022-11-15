package events

import (
	commonDomain "github.com/anthonyraymond/joal-cli/internal/v3/commons/domain"
)

type TorrentPeersChanged struct {
	TorrentId commonDomain.TorrentId
	Peers     []commonDomain.Peers
}

func NewTorrentPeersChanged(torrentId commonDomain.TorrentId, peers []commonDomain.Peers) TorrentPeersChanged {
	return TorrentPeersChanged{TorrentId: torrentId, Peers: peers}
}
