package domain

import (
	"fmt"
	commonDomain "github.com/anthonyraymond/joal-cli/internal/v3/commons/domain"
)

type SharedTorrent struct {
	TorrentId   commonDomain.TorrentId
	Stats       SharingStats
	Peers       commonDomain.Peers
	Downloading bool
	Seeding     bool
}

func (s *SharedTorrent) SetPeers(peers commonDomain.Peers) {
	s.Peers = peers
	if s.Peers.Leechers == 0 {
		s.Stats.Upload.Speed.bps = 0
		s.Stats.Download.Speed.bps = 0
	}
}

func (s *SharedTorrent) AddDownloaded(bytes int64) error {
	if s.Peers.Seeders == 0 {
		return fmt.Errorf("can not add download to a torrent with 0 seeders or 0 leechers")
	}
	return s.Stats.addDownloaded(bytes)
}

func (s *SharedTorrent) AddUploaded(bytes int64) error {
	if s.Peers.Leechers == 0 {
		return fmt.Errorf("can not add upload to a torrent with 0 seeders or 0 leechers")
	}
	return s.Stats.addUploaded(bytes)
}

func (s *SharedTorrent) IsDownloadComplete() bool {
	left := s.Stats.Left.Amount()
	return left.GetInBytes() == 0
}
