package sharing

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTier_shouldFindTrackerFromAnnounceUrl(t *testing.T) {
	ti := newTier([]Tracker{
		fakeTracker("http://localhost:8080"),
		fakeTracker("http://localhost:8081/announce?dqsdq=qq"),
		fakeTracker("http://localhost:8082"),
	})

	found, tracker := ti.search(mustParseUrl("http://localhost:8081/announce?dqsdq=qq"))
	assert.True(t, found, "should have found the tracker")
	assert.Equal(t, ti.trackers[1], tracker)
}

func TestTier_shouldReturnFalseIfTrackerNotFound(t *testing.T) {
	ti := newTier([]Tracker{
		fakeTracker("http://localhost:8080"),
	})

	found, tracker := ti.search(mustParseUrl("http://localhost:9999"))
	assert.False(t, found)
	assert.Nil(t, tracker)
}

func TestTier_shouldDeprioritizeFirstTracker(t *testing.T) {
	ti := newTier([]Tracker{
		fakeTracker("http://localhost:8080"),
		fakeTracker("http://localhost:8081"),
		fakeTracker("http://localhost:8082"),
	})

	ti.deprioritizeTracker(mustParseUrl("http://localhost:8080"))

	assert.Equal(t, newTier([]Tracker{
		fakeTracker("http://localhost:8081"),
		fakeTracker("http://localhost:8082"),
		fakeTracker("http://localhost:8080"),
	}), ti)
}

func TestTier_shouldDeprioritizeLastTracker(t *testing.T) {
	ti := newTier([]Tracker{
		fakeTracker("http://localhost:8080"),
		fakeTracker("http://localhost:8081"),
		fakeTracker("http://localhost:8082"),
	})

	ti.deprioritizeTracker(mustParseUrl("http://localhost:8082"))

	assert.Equal(t, newTier([]Tracker{
		fakeTracker("http://localhost:8080"),
		fakeTracker("http://localhost:8081"),
		fakeTracker("http://localhost:8082"),
	}), ti)
}

func TestTier_shouldDeprioritizeSingleton(t *testing.T) {
	ti := newTier([]Tracker{
		fakeTracker("http://localhost:8080"),
	})

	ti.deprioritizeTracker(mustParseUrl("http://localhost:8082"))

	assert.Equal(t, newTier([]Tracker{
		fakeTracker("http://localhost:8080"),
	}), ti)
}

func TestFindTrackerIndex_shouldFindIndex(t *testing.T) {
	index, err := findTrackerIndex([]Tracker{
		fakeTracker("http://localhost:8080"),
		fakeTracker("http://localhost:8081"),
		fakeTracker("http://localhost:8082"),
	}, mustParseUrl("http://localhost:8080"))

	assert.Equal(t, 0, index)
	assert.NoError(t, err)
}

func TestTier_activeTrackersShouldReturnEmptyIfAllDisabled(t *testing.T) {
	disabledTracker := func(u string) Tracker {
		tr := fakeTracker(u)
		tr.disable(AnnounceProtocolNotSupported)
		return tr
	}

	ti := newTier([]Tracker{
		disabledTracker("http://localhost:8080"),
		disabledTracker("http://localhost:8081"),
	})

	trackers := ti.activeTrackers(allTrackerInTier)

	assert.Empty(t, trackers)
}

func TestTier_activeTrackersShouldIgnoreDisabled(t *testing.T) {
	disabledTracker := func(u string) Tracker {
		tr := fakeTracker(u)
		tr.disable(AnnounceProtocolNotSupported)
		return tr
	}

	ti := newTier([]Tracker{
		fakeTracker("http://localhost:8080"),
		disabledTracker("http://localhost:8081"),
		disabledTracker("http://localhost:8082"),
		fakeTracker("http://localhost:8083"),
	})

	trackers := ti.activeTrackers(allTrackerInTier)

	assert.Equal(t, []Tracker{
		fakeTracker("http://localhost:8080"),
		fakeTracker("http://localhost:8083"),
	}, trackers)
}

func TestTier_activeTrackersWithSelectionModeFirst(t *testing.T) {
	disabledTracker := func(u string) Tracker {
		tr := fakeTracker(u)
		tr.disable(AnnounceProtocolNotSupported)
		return tr
	}

	ti := newTier([]Tracker{
		fakeTracker("http://localhost:8080"),
		disabledTracker("http://localhost:8081"),
		disabledTracker("http://localhost:8082"),
		fakeTracker("http://localhost:8083"),
	})

	trackers := ti.activeTrackers(singleTrackerInTier)

	assert.Equal(t, []Tracker{
		fakeTracker("http://localhost:8080"),
	}, trackers)
}

func TestTrackerSelectionMode_allTrackerInTierShouldReturnAll(t *testing.T) {
	trackers := []Tracker{
		fakeTracker("http://localhost:8080"),
		fakeTracker("http://localhost:8081"),
	}

	assert.Equal(t, trackers, allTrackerInTier(trackers))
}

func TestTrackerSelectionMode_singleTrackerInTierShouldReturnFirst(t *testing.T) {
	trackers := []Tracker{
		fakeTracker("http://localhost:8080"),
		fakeTracker("http://localhost:8081"),
	}

	assert.Equal(t, []Tracker{trackers[0]}, singleTrackerInTier(trackers))
}

func TestTrackerSelectionMode_singleTrackerInTierShouldNotFailWithEmptyTrackerList(t *testing.T) {
	var trackers []Tracker
	assert.Empty(t, singleTrackerInTier(trackers))
}
