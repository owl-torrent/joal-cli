package torrent

import (
	"errors"
)

type linkedTierList struct {
	ITierAnnouncer
	currentIndex uint
	list         []ITierAnnouncer
}

func (l *linkedTierList) next() {
	l.currentIndex = (l.currentIndex + 1) % uint(len(l.list))
	l.ITierAnnouncer = l.list[l.currentIndex]
}

func (l *linkedTierList) rewindToFirst() {
	l.currentIndex = 0
	l.ITierAnnouncer = l.list[l.currentIndex]
}

func (l linkedTierList) isFirst() bool {
	return l.currentIndex == 0
}

func newLinkedTierList(tiers []ITierAnnouncer) (*linkedTierList, error) {
	if len(tiers) == 0 {
		return nil, errors.New("tiers list can not be empty")
	}
	return &linkedTierList{
		ITierAnnouncer: tiers[0],
		currentIndex:   0,
		list:           tiers,
	}, nil
}
