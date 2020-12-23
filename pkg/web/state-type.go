package web

import (
	"github.com/Masterminds/semver/v3"
	"time"
)

type coreState string

const (
	CORE_STARTING coreState = "STARTING"
	CORE_STARTED  coreState = "STARTED"
	CORE_STOPPING coreState = "STARTING"
	CORE_STOPPED  coreState = "STOPPED"
)

type State struct {
	State     coreState          `json:"state"`
	Client    *Client            `json:"client"`
	Config    *Config            `json:"config"`
	Torrents  map[string]Torrent `json:"torrents"`
	Bandwidth *Bandwidth         `json:"bandwidth"`
}

type Client struct {
	Name    string         `json:"name"`
	Version semver.Version `json:"version"`
}

type Config struct {
	NeedRestartToTakeEffect bool           `json:"needRestartToTakeEffect"`
	RuntimeConfig           *RuntimeConfig `json:"runtimeConfig"`
}

type RuntimeConfig struct {
	MinimumBytesPerSeconds int    `json:"minimumBytesPerSeconds"`
	MaximumBytesPerSeconds int    `json:"maximumBytesPerSeconds"`
	Client                 string `json:"client"`
}

type Torrent struct {
	Infohash            string            `json:"infohash"`
	Name                string            `json:"name"`
	File                string            `json:"file"`
	TrackerAnnounceUrls []string          `json:"trackerAnnounceUrls"`
	Size                int64             `json:"size"`
	AnnounceHistory     []IAnnounceResult `json:"announceHistory"`
	Seeders             int               `json:"seeders"`
	Leechers            int               `json:"leechers"`
	Uploaded            int64             `json:"uploaded"`
}

type IAnnounceResult interface {
	TrackerUrl() string
	WasSuccessful() bool
}

type SuccessAnnounceResult struct {
	TrackerUrl    string    `json:"trackerUrl"`
	WasSuccessful bool      `json:"wasSuccessful"`
	Datetime      time.Time `json:"datetime"`
	Seeders       int       `json:"seeders"`
	Leechers      int       `json:"leechers"`
	Interval      int       `json:"interval"`
}

type ErrorAnnounceResult struct {
	TrackerUrl    string    `json:"trackerUrl"`
	WasSuccessful bool      `json:"wasSuccessful"`
	Datetime      time.Time `json:"datetime"`
	Reason        string    `json:"reason"`
}

type Bandwidth struct {
	CurrentBandwidth int64                        `json:"current_bandwidth"`
	Torrents         map[string]*TorrentBandwidth `json:"torrents"`
}

type TorrentBandwidth struct {
	Infohash           string  `json:"infohash"`
	PercentOfBandwidth float32 `json:"percentOfBandwidth"`
}
