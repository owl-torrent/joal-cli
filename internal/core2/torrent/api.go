package torrent

import (
	libtorrent "github.com/anacrolix/torrent"
	libtracker "github.com/anacrolix/torrent/tracker"
	"net/url"
	"time"
)

type Factory interface {
	// CreateOne create a Torrent. announceList might have
	CreateOne(announce, announceList [][]url.URL, announcePolicy AnnouncePolicy) Torrent
}

type Torrent interface {
	InfoHash() libtorrent.InfoHash
	Name() string
	GetPeers() Peers

	// AnnounceStop unconditionally send Stop event to all tracker currently in use
	AnnounceStop(event libtracker.AnnounceEvent, announcingFunction AnnouncingFunction)
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
