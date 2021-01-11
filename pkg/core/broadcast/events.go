package broadcast

import (
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/core/config"
	"net/url"
	"time"
)

type SeedStartedEvent struct {
	Client  string
	Version string
}

type SeedStoppedEvent struct {
}

type ConfigChangedEvent struct {
	NeedRestartToTakeEffect bool
	RuntimeConfig           *config.RuntimeConfig
}

type TorrentAddedEvent struct {
	Infohash            torrent.InfoHash
	Name                string
	File                string
	TrackerAnnounceUrls []*url.URL
	Size                int64
}

type TorrentAnnouncingEvent struct {
	Infohash      torrent.InfoHash
	TrackerUrl    url.URL
	AnnounceEvent tracker.AnnounceEvent
	Uploaded      int64
}

type TorrentAnnounceSuccessEvent struct {
	Infohash      torrent.InfoHash
	TrackerUrl    url.URL
	AnnounceEvent tracker.AnnounceEvent
	Datetime      time.Time
	Seeder        int32
	Leechers      int32
	Interval      time.Duration
}

type TorrentAnnounceFailedEvent struct {
	Infohash      torrent.InfoHash
	TrackerUrl    url.URL
	AnnounceEvent tracker.AnnounceEvent
	Datetime      time.Time
	Error         string
}

type TorrentSwarmChangedEvent struct {
	Infohash torrent.InfoHash
	Seeder   int32
	Leechers int32
}

type TorrentRemovedEvent struct {
	Infohash torrent.InfoHash
}

type NoticeableErrorEvent struct {
	Error    error
	Datetime time.Time
}

type GlobalBandwidthChangedEvent struct {
	AvailableBandwidth int64
}

type BandwidthWeightHasChangedEvent struct {
	TotalWeight    float64
	TorrentWeights map[torrent.InfoHash]float64
}
