package broadcast

import (
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/internal/config"
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
	TrackerAnnounceUrls []string
	Size                int64
}

type TorrentAnnouncingEvent struct {
	Infohash      torrent.InfoHash
	TrackerUrl    string
	announceEvent tracker.AnnounceEvent
}

type TorrentAnnounceSuccessEvent struct {
	Infohash      torrent.InfoHash
	TrackerUrl    string
	announceEvent tracker.AnnounceEvent
	Datetime      time.Time
	Seeder        int
	Leechers      int
	Interval      time.Duration
}

type TorrentAnnounceFailedEvent struct {
	Infohash      torrent.InfoHash
	TrackerUrl    string
	announceEvent tracker.AnnounceEvent
	Datetime      time.Time
	Error         string
}

type TorrentSwarmChangedEvent struct {
	Infohash torrent.InfoHash
	Seeder   int
	Leechers int
}

type TorrentRemovedEvent struct {
	Infohash torrent.InfoHash
}

type NoticeableErrorEvent struct {
	Error error
}

type GlobalBandwidthChangedEvent struct {
	AvailableBandwidth int64
}

type BandwidthWeightHasChangedEvent struct {
	TotalWeight    int64
	TorrentWeights map[torrent.InfoHash]int64
}
