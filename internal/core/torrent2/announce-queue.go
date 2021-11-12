package torrent2

import (
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/tracker"
	"net/url"
	"time"
)

const queueCapacity int = 1500

type AnnounceQueue struct {
	queue chan *AnnounceRequest
}

func New() *AnnounceQueue {
	return &AnnounceQueue{
		queue: make(chan *AnnounceRequest, queueCapacity),
	}
}

func (q *AnnounceQueue) Enqueue(req *AnnounceRequest) {
	q.queue <- req
}

func (q *AnnounceQueue) Request() <-chan *AnnounceRequest {
	return q.queue
}

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
	Request  *AnnounceRequest
	Interval time.Duration // Minimum seconds the local peer should wait before next announce. May be 0 if the error is not related to the tracker response
}

type AnnounceCallbacks struct {
	Success func(AnnounceResponse)
	Failed  func(AnnounceResponseError)
}
