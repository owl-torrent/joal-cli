package common

import (
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/tracker"
	"net/url"
	"time"
)

type AnnounceRequest struct {
	//TODO: not sure if needed so far => Ctx context.Context
	Url        url.URL
	InfoHash   torrent.InfoHash
	Downloaded int64
	Left       int64
	Uploaded   int64
	Corrupt    int64
	Event      tracker.AnnounceEvent
	Private    bool
	// TODO: not sure about impl at this point => AnnounceCallbacks AnnounceCallbacks
}

type AnnounceResponseError struct {
	Error   error
	Request AnnounceRequest
}
type AnnounceResponse struct {
	Request  AnnounceRequest
	Interval time.Duration
	Leechers int32
	Seeders  int32
}
