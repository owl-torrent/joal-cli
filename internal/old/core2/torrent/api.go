package torrent

import (
	libtorrent "github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"time"
)

type Factory interface {
	// CreateTorrent create a Torrent. announceList might have been shuffled
	CreateTorrent(meta metainfo.MetaInfo, announcePolicy AnnouncePolicy) (Torrent, error)
}

type Torrent interface {
	InfoHash() libtorrent.InfoHash
	Name() string
	GetPeers() Peers

	// AnnounceStop unconditionally send Stop event to all tracker currently in use
	AnnounceStop(announcingFunction AnnouncingFunction)
	// AnnounceToReadyTrackers announce to all tracker that are ready to receive an announce
	AnnounceToReadyTrackers(announcingFunction AnnouncingFunction)
	// HandleAnnounceSuccess delegate the handling of the response to the trackers
	HandleAnnounceSuccess(response TrackerAnnounceResponse)
	// HandleAnnounceError delegate the handling of the response to the trackers
	HandleAnnounceError(response TrackerAnnounceResponseError)
}

type TrackerAnnounceResponse struct {
	Request  TrackerAnnounceRequest
	Interval time.Duration
	Seeders  int32
	Leechers int32
}

type TrackerAnnounceResponseError struct {
	Error   error
	Request TrackerAnnounceRequest
}

type Peers interface {
	Seeders() int32
	Leechers() int32
}
