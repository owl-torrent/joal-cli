package events

import (
	v3 "github.com/anthonyraymond/joal-cli/internal/old/v3/commons/domain"
)

type TorrentDownloaded struct {
	TorrentId v3.TorrentId
}

func NewTorrentDownloaded(torrentId v3.TorrentId) TorrentDownloaded {
	return TorrentDownloaded{TorrentId: torrentId}
}
