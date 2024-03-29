package announces

import (
	"context"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/tracker"
	"net/url"
	"time"
)

type AnnounceRequest struct {
	Ctx               context.Context
	Url               url.URL
	InfoHash          torrent.InfoHash
	Downloaded        int64
	Left              int64
	Uploaded          int64
	Corrupt           int64
	Event             tracker.AnnounceEvent
	Private           bool
	AnnounceCallbacks *AnnounceCallbacks
}

type AnnounceResponse struct {
	Request  *AnnounceRequest
	Interval time.Duration // Minimum seconds the local peer should wait before next announce.
	Leechers int32
	Seeders  int32
	Peers    []tracker.Peer
}

type AnnounceResponseError struct {
	Error    error
	Request  *AnnounceRequest
	Interval time.Duration // Minimum seconds the local peer should wait before next announce. May be 0 if the error is not related to the tracker response
}

type AnnounceCallbacks struct {
	Success func(AnnounceResponse)
	Failed  func(AnnounceResponseError)
}
