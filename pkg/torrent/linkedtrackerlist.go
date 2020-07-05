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

	selected := l.list[l.currentIndex]
	for i := l.currentIndex; i > 0; i-- { // offset by +1 all the element before the current
		l.list[i] = l.list[i-1]
	}
	l.list[0] = selected

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

func (l *linkedTrackerList) isLast() bool {
	l.lock.RLock()
	r := l.currentIndex == uint(len(l.list)-1)
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
