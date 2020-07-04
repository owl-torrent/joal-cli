package torrent

import (
	"errors"
	"sync"
)

type linkedTrackerList struct {
	ITrackerAnnouncer
	currentIndex uint
	list         []ITrackerAnnouncer
	lock         *sync.RWMutex
}

func (l *linkedTrackerList) next() {
	l.lock.Lock()
	l.currentIndex = (l.currentIndex + 1) % uint(len(l.list))
	l.ITrackerAnnouncer = l.list[l.currentIndex]
	l.lock.Unlock()
}

func (l *linkedTrackerList) PromoteCurrent() {
	l.lock.Lock()
	if l.currentIndex == 0 {
		l.lock.Unlock()
		return
	}

	for i := l.currentIndex; i > 0; i-- {
		l.list[i] = l.list[i-1]
	}

	l.currentIndex = 0
	l.ITrackerAnnouncer = l.list[l.currentIndex]
	l.lock.Unlock()
}

func (l *linkedTrackerList) isFirst() bool {
	l.lock.RLock()
	r := l.currentIndex == 0
	l.lock.RUnlock()
	return r
}

func newLinkedTrackerList(tiers []ITrackerAnnouncer) (*linkedTrackerList, error) {
	if len(tiers) == 0 {
		return nil, errors.New("tiers list can not be empty")
	}
	return &linkedTrackerList{
		ITrackerAnnouncer: tiers[0],
		currentIndex:      0,
		list:              tiers,
		lock:              &sync.RWMutex{},
	}, nil
}
