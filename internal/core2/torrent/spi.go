package torrent

import (
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
	Event      libtracker.AnnounceEvent
	Url        url.URL
	Uploaded   int64
	Downloaded int64
	Left       int64
	Corrupt    int64
}

type AnnouncingFunction = func(TrackerAnnounceRequest)
