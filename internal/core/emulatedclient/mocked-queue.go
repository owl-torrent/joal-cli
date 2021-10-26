package emulatedclient

import (
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/tracker"
	"net/url"
	"time"
)

type AnnounceRequest struct {
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
	error
	Request *AnnounceRequest
}

type AnnounceCallbacks struct {
	Success func(AnnounceResponse)
	Failed  func(AnnounceResponseError)
}
