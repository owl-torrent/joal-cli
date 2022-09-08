package torrent

import (
	libtracker "github.com/anacrolix/torrent/tracker"
	"time"
)

type AnnounceAbleTorrent interface {
	// AnnounceStop unconditionally send Stop event to all tracker currently in use
	AnnounceStop(event libtracker.AnnounceEvent, announcingFunction AnnouncingFunction)
	// AnnounceToReadyTrackers announce to all tracker that are ready to receive an announce
	AnnounceToReadyTrackers(announcingFunction AnnouncingFunction)
	// HandleAnnounceSuccess delegate the handling of the response to the trackers
	HandleAnnounceSuccess(response TrackerAnnounceResponse)
	// HandleAnnounceError delegate the handling of the response to the trackers
	HandleAnnounceError(response TrackerAnnounceResponseError)

	GetPeers() Peers
}

type TrackerAnnounceResponse struct {
	Request  TrackerAnnounceRequest
	Interval time.Duration
	Leechers int32
	Seeders  int32
}

type TrackerAnnounceResponseError struct {
	Error   error
	Request TrackerAnnounceRequest
}

type Peers interface {
	Seeders() int32
	Leechers() int32
}
