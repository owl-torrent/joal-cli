package web

import (
	"github.com/anacrolix/torrent/tracker"
	"net/url"
	"time"
)

type State struct {
	Started   bool                `json:"started"`
	Client    *Client             `json:"client"`
	Config    *Config             `json:"config"`
	Torrents  map[string]*Torrent `json:"torrents"`
	Bandwidth *Bandwidth          `json:"bandwidth"`
}

func (s State) InitialState() *State {
	return &State{
		Started: false,
	}
}

type Client struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Config struct {
	NeedRestartToTakeEffect bool           `json:"needRestartToTakeEffect"`
	RuntimeConfig           *RuntimeConfig `json:"runtimeConfig"`
}

type RuntimeConfig struct {
	MinimumBytesPerSeconds int64  `json:"minimumBytesPerSeconds"`
	MaximumBytesPerSeconds int64  `json:"maximumBytesPerSeconds"`
	Client                 string `json:"client"`
}

type Torrent struct {
	Infohash string                      `json:"infohash"`
	Name     string                      `json:"name"`
	File     string                      `json:"file"`
	Size     int64                       `json:"size"`
	Seeders  int32                       `json:"seeders"`
	Leechers int32                       `json:"leechers"`
	Uploaded int64                       `json:"uploaded"`
	Trackers map[string]*TorrentTrackers `json:"trackers"`
}

type TorrentTrackers struct {
	Url             *url.URL          `json:"url"`
	IsAnnouncing    bool              `json:"isAnnouncing"`
	InUse           bool              `json:"inUse"`
	Seeders         int32             `json:"seeders"`
	Leechers        int32             `json:"leechers"`
	Interval        int               `json:"interval"`
	AnnounceHistory []*AnnounceResult `json:"announceHistory"`
}

type AnnounceResult struct {
	AnnounceEvent tracker.AnnounceEvent `json:"announceEvent"`
	WasSuccessful bool                  `json:"wasSuccessful"`
	Datetime      time.Time             `json:"datetime"`
	Seeders       int32                 `json:"seeders"`
	Leechers      int32                 `json:"leechers"`
	Interval      int                   `json:"interval"`
	Error         string                `json:"reason,omitempty"`
}

type Bandwidth struct {
	CurrentBandwidth int64                        `json:"current_bandwidth"`
	Torrents         map[string]*TorrentBandwidth `json:"torrents"`
}

type TorrentBandwidth struct {
	Infohash           string  `json:"infohash"`
	PercentOfBandwidth float32 `json:"percentOfBandwidth"`
}
