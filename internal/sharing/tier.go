package sharing

import (
	"errors"
	"net/url"
	"strings"
)

type trackersSelectionMode func([]Tracker) []Tracker

var (
	allTrackerInTier trackersSelectionMode = func(input []Tracker) []Tracker {
		return input
	}
	singleTrackerInTier trackersSelectionMode = func(input []Tracker) []Tracker {
		if len(input) == 0 {
			return input
		}
		return input[:1]
	}
)

type tier struct {
	trackers []Tracker
}

func (t *tier) search(u *url.URL) (bool, Tracker) {
	index, err := findTrackerIndex(t.trackers, u)
	if err != nil {
		return false, nil
	}
	return true, t.trackers[index]
}

func (t *tier) deprioritizeTracker(u *url.URL) {
	index, err := findTrackerIndex(t.trackers, u)
	if err != nil {
		return
	}

	temp := t.trackers[index]

	for i := index; i < len(t.trackers)-1; i++ {
		t.trackers[i] = t.trackers[i+1]
	}
	t.trackers[len(t.trackers)-1] = temp
}

func (t *tier) activeTrackers(selectionMode trackersSelectionMode) []Tracker {
	var active []Tracker

	for i := range t.trackers {
		if !t.trackers[i].IsDisabled() {
			active = append(active, t.trackers[i])
		}
	}

	return selectionMode(active)
}

func newTier(trackers []Tracker) *tier {
	return &tier{
		trackers: trackers,
	}
}

func findTrackerIndex(trackers []Tracker, searchUrl *url.URL) (int, error) {
	for i := range trackers {
		if strings.EqualFold(trackers[i].Url().String(), searchUrl.String()) {
			return i, nil
		}
	}

	return -1, errors.New("tracker not found")
}
