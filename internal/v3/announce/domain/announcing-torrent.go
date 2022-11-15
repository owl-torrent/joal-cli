package domain

import commonDomain "github.com/anthonyraymond/joal-cli/internal/v3/commons/domain"

type Torrent struct {
	TorrentId commonDomain.TorrentId
	Trackers  Trackers
}
