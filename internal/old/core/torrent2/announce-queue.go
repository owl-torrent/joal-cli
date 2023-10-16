package torrent2

import (
	"github.com/anthonyraymond/joal-cli/internal/old/core/announces"
	"sync"
)

const queueCapacity int = 1500

type AnnounceQueue struct {
	queue    chan *announces.AnnounceRequest
	isClosed bool
	lock     *sync.RWMutex
}

func NewAnnounceQueue() *AnnounceQueue {
	return &AnnounceQueue{
		queue:    make(chan *announces.AnnounceRequest, queueCapacity),
		isClosed: false,
		lock:     &sync.RWMutex{},
	}
}

func (q *AnnounceQueue) Enqueue(req *announces.AnnounceRequest) {
	q.lock.RLock()
	if q.isClosed {
		q.lock.RUnlock()
		return
	}
	q.lock.RUnlock()

	q.queue <- req
}

func (q *AnnounceQueue) Request() <-chan *announces.AnnounceRequest {
	return q.queue
}

func (q *AnnounceQueue) DiscardFutureEnqueueAndDestroy() {
	q.lock.Lock()
	if q.isClosed {
		q.lock.Unlock()
		return
	}

	q.isClosed = true
	close(q.queue)

	q.lock.Unlock()
}
