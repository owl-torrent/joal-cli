package torrent

import (
	libtorrent "github.com/anacrolix/torrent"
	libtracker "github.com/anacrolix/torrent/tracker"
	"net/url"
)

type AnnouncePolicy interface {
	SupportHttpAnnounce() bool
	SupportUdpAnnounce() bool
	SupportAnnounceList() bool
	ShouldAnnounceToAllTier() bool
	ShouldAnnounceToAllTrackersInTier() bool
}

type TrackerAnnounceRequest struct {
	InfoHash   libtorrent.InfoHash
	Event      libtracker.AnnounceEvent
	Url        url.URL
	Uploaded   int64
	Downloaded int64
	Left       int64
	Corrupt    int64
	Private    bool
}

type AnnouncingFunction = func(TrackerAnnounceRequest)
