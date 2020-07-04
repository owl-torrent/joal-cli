package torrent

import (
	"errors"
	"sync"
)

type linkedTierList struct {
	ITierAnnouncer
	currentIndex uint
	list         []ITierAnnouncer
	lock         *sync.RWMutex
}

func (l *linkedTierList) next() {
	l.lock.Lock()
	l.currentIndex = (l.currentIndex + 1) % uint(len(l.list))
	l.ITierAnnouncer = l.list[l.currentIndex]
	l.lock.Unlock()
}

func (l *linkedTierList) backToFirst() {
	l.lock.Lock()
	l.currentIndex = 0
	l.ITierAnnouncer = l.list[l.currentIndex]
	l.lock.Unlock()
}

func (l linkedTierList) isFirst() bool {
	l.lock.RLock()
	r := l.currentIndex == 0
	l.lock.RUnlock()
	return r
}

func newLinkedTierList(tiers []ITierAnnouncer) (*linkedTierList, error) {
	if len(tiers) == 0 {
		return nil, errors.New("tiers list can not be empty")
	}
	return &linkedTierList{
		ITierAnnouncer: tiers[0],
		currentIndex:   0,
		list:           tiers,
		lock:           &sync.RWMutex{},
	}, nil
}
