package web

import (
	"net/url"
	"time"
)

type state struct {
	Global    *globalState             `json:"global"`
	Config    *configState             `json:"config"`
	Torrents  map[string]*torrentState `json:"torrents"`
	Bandwidth *bandwidthState          `json:"bandwidth"`
}

func (s state) initialState() *state {
	return &state{
		Global: &globalState{
			Started: false,
		},
	}
}

type clientState struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type globalState struct {
	Started bool         `json:"started"`
	Client  *clientState `json:"client"`
}

type configState struct {
	NeedRestartToTakeEffect bool                `json:"needRestartToTakeEffect"`
	RuntimeConfig           *runtimeConfigState `json:"runtimeConfig"`
}

type runtimeConfigState struct {
	MinimumBytesPerSeconds int64  `json:"minimumBytesPerSeconds"`
	MaximumBytesPerSeconds int64  `json:"maximumBytesPerSeconds"`
	Client                 string `json:"client"`
}

type torrentState struct {
	Infohash string                           `json:"infohash"`
	Name     string                           `json:"name"`
	File     string                           `json:"file"`
	Size     int64                            `json:"size"`
	Seeders  int32                            `json:"seeders"`
	Leechers int32                            `json:"leechers"`
	Uploaded int64                            `json:"uploaded"`
	Trackers map[string]*torrentTrackersState `json:"trackers"`
}

type torrentTrackersState struct {
	Url             *url.URL               `json:"url"`
	IsAnnouncing    bool                   `json:"isAnnouncing"`
	InUse           bool                   `json:"inUse"`
	Seeders         int32                  `json:"seeders"`
	Leechers        int32                  `json:"leechers"`
	Interval        int                    `json:"interval"`
	AnnounceHistory []*announceResultState `json:"announceHistory"`
}

type announceResultState struct {
	AnnounceEvent string    `json:"announceEvent"`
	WasSuccessful bool      `json:"wasSuccessful"`
	Datetime      time.Time `json:"datetime"`
	Seeders       int32     `json:"seeders"`
	Leechers      int32     `json:"leechers"`
	Interval      int       `json:"interval"`
	Error         string    `json:"reason,omitempty"`
}

type bandwidthState struct {
	CurrentBandwidth int64                             `json:"currentBandwidth"`
	Torrents         map[string]*torrentBandwidthState `json:"torrents"`
}

type torrentBandwidthState struct {
	Infohash           string  `json:"infohash"`
	PercentOfBandwidth float32 `json:"percentOfBandwidth"`
}
