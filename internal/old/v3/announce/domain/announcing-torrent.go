package domain

import commonDomain "github.com/anthonyraymond/joal-cli/internal/old/v3/commons/domain"

type AnnouncingTorrent struct {
	TorrentId commonDomain.TorrentId
	Trackers  Trackers
}
