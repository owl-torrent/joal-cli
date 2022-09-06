package torrent

import (
	"fmt"
	"github.com/anthonyraymond/joal-cli/internal/core2/common"
	"github.com/anthonyraymond/joal-cli/internal/utils/testutils"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

type testAnnouncePolicy struct {
	supportHttpAnnounce               bool
	supportUdpAnnounce                bool
	supportAnnounceList               bool
	shouldAnnounceToAllTier           bool
	shouldAnnounceToAllTrackersInTier bool
}

func (t testAnnouncePolicy) SupportHttpAnnounce() bool {
	return t.supportHttpAnnounce
}

func (t testAnnouncePolicy) SupportUdpAnnounce() bool {
	return t.supportUdpAnnounce
}

func (t testAnnouncePolicy) SupportAnnounceList() bool {
	return t.supportAnnounceList
}

func (t testAnnouncePolicy) ShouldAnnounceToAllTier() bool {
	return t.shouldAnnounceToAllTier
}

func (t testAnnouncePolicy) ShouldAnnounceToAllTrackersInTier() bool {
	return t.shouldAnnounceToAllTrackersInTier
}

func TestCreateTrackers(t *testing.T) {
	announcePolicy := testAnnouncePolicy{
		supportHttpAnnounce: true,
		supportUdpAnnounce:  true,
		supportAnnounceList: true,
	}

	announceList := [][]string{
		{"udp://localhost:20/tier1/tr1", "http://localhost:20/tier1/tr2"},
		{"http://localhost:20/tier2/tr1", "https://localhost:20/tier2/tr2", "udp://localhost:20/tier2/tr3"},
	}

	trackers, err := CreateTrackers("http://localhost:1111/announce", announceList, announcePolicy)
	assert.NoError(t, err)

	for _, tr := range trackers.trackers {
		assert.NotEqual(t, "http://localhost:1111/announce", tr.url.String(), "when supportAnnounceList is true, the single announce url should not be added to tracker list")
		assert.False(t, tr.disabled.IsDisabled(), "no trackers should be disabled")
	}
	assert.Len(t, trackers.trackers, 5)
}

func TestCreateTrackers_ShouldShuffleTrackersInAllAnnounceListTiers(t *testing.T) {
	announcePolicy := testAnnouncePolicy{
		supportHttpAnnounce: true,
		supportUdpAnnounce:  true,
		supportAnnounceList: true,
	}

	announceList := [][]string{
		{"udp://localhost:20/tier1/tr1", "http://localhost:20/tier1/tr2", "http://localhost:20/tier1/tr3", "http://localhost:20/tier1/tr4", "http://localhost:20/tier1/tr5"},
		{"http://localhost:20/tier2/tr1", "https://localhost:20/tier2/tr2", "udp://localhost:20/tier2/tr3", "udp://localhost:20/tier2/tr4", "udp://localhost:20/tier2/tr5"},
	}

	trackers, err := CreateTrackers("http://localhost:1111/announce", announceList, announcePolicy)
	assert.NoError(t, err)

	//size should not have changed
	assert.Len(t, trackers.trackers, 10)

	// ensure shuffled
	currentAnnounceList := [][]string{{}, {}}
	for _, tr := range trackers.trackers {
		currentAnnounceList[tr.tier-1] = append(currentAnnounceList[tr.tier-1], tr.url.String())
	}
	orderChanged := false
	for tierIdx := range announceList {
		for trackerIdx := range announceList[tierIdx] {
			if announceList[tierIdx][trackerIdx] != currentAnnounceList[tierIdx][trackerIdx] {
				orderChanged = true
			}
		}
	}

	assert.True(t, orderChanged)
}

func TestCreateTrackers_shouldDisableUdpTrackers(t *testing.T) {
	announcePolicy := testAnnouncePolicy{
		supportHttpAnnounce: true,
		supportUdpAnnounce:  false,
		supportAnnounceList: true,
	}

	announceList := [][]string{
		{"udp://localhost:20/tier1/tr1", "http://localhost:20/tier1/tr2"},
		{"http://localhost:20/tier2/tr1", "https://localhost:20/tier2/tr2", "UDP://localhost:20/tier2/tr3"},
	}

	trackers, err := CreateTrackers("http://localhost:1111/announce", announceList, announcePolicy)
	assert.NoError(t, err)

	disabledCount := 0
	for _, tr := range trackers.trackers {
		if strings.HasPrefix(strings.ToLower(tr.url.Scheme), "udp") {
			disabledCount++
			assert.True(t, tr.disabled.IsDisabled())
			assert.Equal(t, announceProtocolNotSupported, tr.disabled)
		}
	}
	assert.Equal(t, 2, disabledCount)
}

func TestCreateTrackers_shouldDisableHttpAndHttpsTrackers(t *testing.T) {
	announcePolicy := testAnnouncePolicy{
		supportHttpAnnounce: false,
		supportUdpAnnounce:  true,
		supportAnnounceList: true,
	}

	announceList := [][]string{
		{"udp://localhost:20/tier1/tr1", "http://localhost:20/tier1/tr2", "HTTP://localhost:20/tier1/tr3"},
		{"http://localhost:20/tier2/tr1", "https://localhost:20/tier2/tr2", "udp://localhost:20/tier2/tr3"},
	}

	trackers, err := CreateTrackers("http://localhost:1111/announce", announceList, announcePolicy)
	assert.NoError(t, err)

	disabledCount := 0
	for _, tr := range trackers.trackers {
		if strings.HasPrefix(strings.ToLower(tr.url.Scheme), "http") {
			disabledCount++
			assert.True(t, tr.disabled.IsDisabled())
			assert.Equal(t, announceProtocolNotSupported, tr.disabled)
		}
	}
	assert.Equal(t, 4, disabledCount)
}

func TestCreateTrackers_shouldDisableUnknownTrackers(t *testing.T) {
	announcePolicy := testAnnouncePolicy{
		supportHttpAnnounce: true,
		supportUdpAnnounce:  true,
		supportAnnounceList: true,
	}

	announceList := [][]string{
		{"udp://localhost:20/tier1/tr1", "dop://localhost:20/tier1/tr2"},
		{"http://localhost:20/tier2/tr1", "pat://localhost:20/tier2/tr2", "udp://localhost:20/tier2/tr3"},
	}

	trackers, err := CreateTrackers("http://localhost:1111/announce", announceList, announcePolicy)
	assert.NoError(t, err)

	disabledCount := 0
	for _, tr := range trackers.trackers {
		if strings.HasPrefix(strings.ToLower(tr.url.Scheme), "dop") || strings.HasPrefix(strings.ToLower(tr.url.Scheme), "pat") {
			disabledCount++
			assert.True(t, tr.disabled.IsDisabled())
			assert.Equal(t, announceProtocolNotSupported, tr.disabled)
		}
	}
	assert.Equal(t, 2, disabledCount)
}

func TestCreateTrackers_shouldReturnErrorIfAllTrackersAreDisabled(t *testing.T) {
	announcePolicy := testAnnouncePolicy{
		supportHttpAnnounce: false,
		supportUdpAnnounce:  false,
		supportAnnounceList: true,
	}

	announceList := [][]string{
		{"udp://localhost:20/tier1/tr1", "http://localhost:20/tier1/tr2"},
		{"http://localhost:20/tier2/tr1", "https://localhost:20/tier2/tr2", "udp://localhost:20/tier2/tr3"},
	}

	_, err := CreateTrackers("http://localhost:1111/announce", announceList, announcePolicy)

	assert.Error(t, err)
	assert.Equal(t, ErrAllTrackerAreDisabled, err)
}

func TestCreateTrackers_shouldDisableAllTrackerIfAnnounceListIsNotSupportedAndAddTheSingleAnnounceAsTier0(t *testing.T) {
	announcePolicy := testAnnouncePolicy{
		supportHttpAnnounce: true,
		supportUdpAnnounce:  true,
		supportAnnounceList: false,
	}

	announceList := [][]string{
		{"udp://localhost:20/tier1/tr1", "http://localhost:20/tier1/tr2"},
		{"http://localhost:20/tier2/tr1", "https://localhost:20/tier2/tr2", "udp://localhost:20/tier2/tr3"},
	}

	trackers, err := CreateTrackers("http://localhost:1111/announce", announceList, announcePolicy)
	assert.NoError(t, err)

	// When supportAnnounceList is false, the single "announce" tracker should be moved on alone on tier 1 and all tiers of announceList must be incremented by one
	assert.Len(t, trackers.trackers, 6, "there should be 6 trackers: 1 single announce + 5 of announce list")
	trackersOfTierZero := []Tracker{}
	for _, tracker := range trackers.trackers {
		if tracker.tier == 0 {
			trackersOfTierZero = append(trackersOfTierZero, tracker)
		}
	}
	assert.Len(t, trackersOfTierZero, 1, "Tier zero should contains only the single announce")
	assert.Equal(t, "http://localhost:1111/announce", trackersOfTierZero[0].url.String())

	// All trackers from announceList should be disabled
	disabledCount := 0
	for _, tr := range trackers.trackers {
		if tr.url.String() != "http://localhost:1111/announce" {
			assert.True(t, tr.disabled.IsDisabled())
			assert.Equal(t, announceListNotSupported, tr.disabled)
			disabledCount++
		}
	}
	assert.Equal(t, 5, disabledCount)
}

func TestCreateTrackers_shouldRemoveSingleTrackerUrlFromAnnounceListIfAnnounceListIsNotSupportedAndTheSingleTrackerAlsoPresentInAnnounceList(t *testing.T) {
	announcePolicy := testAnnouncePolicy{
		supportHttpAnnounce: true,
		supportUdpAnnounce:  true,
		supportAnnounceList: false,
	}

	announceList := [][]string{
		{"udp://localhost:20/tier1/tr1", "http://localhost:1111/announce"},
		{"http://localhost:20/tier2/tr1", "http://localhost:1111/announce", "udp://localhost:20/tier2/tr3"},
	}

	trackers, err := CreateTrackers("http://localhost:1111/announce", announceList, announcePolicy)
	assert.NoError(t, err)

	numberOfTimeFound := 0
	for _, tracker := range trackers.trackers {
		if tracker.url.String() == "http://localhost:1111/announce" {
			numberOfTimeFound++
		}
	}
	assert.Equal(t, 1, numberOfTimeFound)
}

func TestCreateTrackers_shouldUseSingleTrackerFromAnnounceListIfSupportAnnounceListIsTrueButAnnounceListIsEmpty(t *testing.T) {
	announcePolicy := testAnnouncePolicy{
		supportHttpAnnounce: true,
		supportUdpAnnounce:  true,
		supportAnnounceList: false,
	}

	announceList := [][]string{{}, {}}

	trackers, err := CreateTrackers("http://localhost:1111/announce", announceList, announcePolicy)
	assert.NoError(t, err)

	assert.Len(t, trackers.trackers, 1)
	assert.Equal(t, trackers.trackers[0].url.String(), "http://localhost:1111/announce")
}

func Test_hasOneEnabled(t *testing.T) {
	var createDisabled = func() Tracker {
		return Tracker{disabled: TrackerDisabled{disabled: true}}
	}
	var createEnabled = func() Tracker {
		return Tracker{disabled: TrackerDisabled{disabled: false}}
	}
	type args struct {
		trackers []Tracker
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "empty should not return enabled", args: args{trackers: []Tracker{}}, want: false},
		{name: "only one but disabled", args: args{trackers: []Tracker{createDisabled()}}, want: false},
		{name: "all disabled", args: args{trackers: []Tracker{createDisabled(), createDisabled()}}, want: false},
		{name: "some enabled", args: args{trackers: []Tracker{createDisabled(), createEnabled()}}, want: true},
		{name: "all enabled", args: args{trackers: []Tracker{createEnabled(), createEnabled()}}, want: true},
		{name: "only one and enabled", args: args{trackers: []Tracker{createEnabled()}}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, hasOneEnabled(tt.args.trackers), "hasOneEnabled(%v)", tt.args.trackers)
		})
	}
}

func Test_findTrackerForUrl(t *testing.T) {
	trackers := []Tracker{
		{url: *testutils.MustParseUrl("http://localhost:1234/announce")},
		{url: *testutils.MustParseUrl("http://localhost:5678/announce")},
	}

	index, err := findTrackerForUrl(trackers, *testutils.MustParseUrl("http://localhost:5678/announce"))
	if err != nil {
		return
	}

	assert.Equal(t, 1, index)
}

func Test_findTrackerForUrlWithoutCaseSensitivity(t *testing.T) {
	trackers := []Tracker{
		{url: *testutils.MustParseUrl("http://localhost:1234/announce")},
		{url: *testutils.MustParseUrl("http://LOCAlhost:5678/announce")},
	}

	index, err := findTrackerForUrl(trackers, *testutils.MustParseUrl("http://localhost:5678/ANNOUnce"))
	if err != nil {
		return
	}

	assert.Equal(t, 1, index)
}

func Test_findTrackerForUrl_ShouldReturnErrorIfUrlIsNotInSlice(t *testing.T) {
	trackers := []Tracker{
		{url: *testutils.MustParseUrl("http://localhost:1234/announce")},
		{url: *testutils.MustParseUrl("http://localhost:5678/announce")},
	}

	_, err := findTrackerForUrl(trackers, *testutils.MustParseUrl("http://localhost:9876/announce"))
	assert.Error(t, err)
}

func Test_deprioritizeTracker(t *testing.T) {
	type args struct {
		trackers            []Tracker
		indexToDeprioritize int
	}
	tests := []struct {
		name string
		args args
		want []Tracker
	}{
		{
			name: "should move tracker 0 last of his tier",
			args: args{
				trackers: []Tracker{
					{url: *testutils.MustParseUrl("http://localhost:8080"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:1236"), tier: 1},
					{url: *testutils.MustParseUrl("http://127.0.0.1:8080"), tier: 2},
					{url: *testutils.MustParseUrl("http://127.0.0.1:4456"), tier: 2},
				},
				indexToDeprioritize: 0,
			},
			want: []Tracker{
				{url: *testutils.MustParseUrl("http://localhost:1236"), tier: 1},
				{url: *testutils.MustParseUrl("http://localhost:8080"), tier: 1},
				{url: *testutils.MustParseUrl("http://127.0.0.1:8080"), tier: 2},
				{url: *testutils.MustParseUrl("http://127.0.0.1:4456"), tier: 2},
			},
		},
		{
			name: "should move tracker 1 last of his tier",
			args: args{
				trackers: []Tracker{
					{url: *testutils.MustParseUrl("http://localhost:8080"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:6549"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:1236"), tier: 1},
					{url: *testutils.MustParseUrl("http://127.0.0.1:8080"), tier: 2},
					{url: *testutils.MustParseUrl("http://127.0.0.1:4456"), tier: 2},
				},
				indexToDeprioritize: 1,
			},
			want: []Tracker{
				{url: *testutils.MustParseUrl("http://localhost:8080"), tier: 1},
				{url: *testutils.MustParseUrl("http://localhost:1236"), tier: 1},
				{url: *testutils.MustParseUrl("http://localhost:6549"), tier: 1},
				{url: *testutils.MustParseUrl("http://127.0.0.1:8080"), tier: 2},
				{url: *testutils.MustParseUrl("http://127.0.0.1:4456"), tier: 2},
			},
		},
		{
			name: "should move tracker 3 last of his tier",
			args: args{
				trackers: []Tracker{
					{url: *testutils.MustParseUrl("http://localhost:8080"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:6549"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:1236"), tier: 1},
					{url: *testutils.MustParseUrl("http://127.0.0.1:8080"), tier: 2},
					{url: *testutils.MustParseUrl("http://127.0.0.1:4456"), tier: 2},
				},
				indexToDeprioritize: 3,
			},
			want: []Tracker{
				{url: *testutils.MustParseUrl("http://localhost:8080"), tier: 1},
				{url: *testutils.MustParseUrl("http://localhost:6549"), tier: 1},
				{url: *testutils.MustParseUrl("http://localhost:1236"), tier: 1},
				{url: *testutils.MustParseUrl("http://127.0.0.1:4456"), tier: 2},
				{url: *testutils.MustParseUrl("http://127.0.0.1:8080"), tier: 2},
			},
		},
		{
			name: "should move tracker last tracker last of his tier",
			args: args{
				trackers: []Tracker{
					{url: *testutils.MustParseUrl("http://localhost:8080"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:6549"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:1236"), tier: 1},
					{url: *testutils.MustParseUrl("http://127.0.0.1:8080"), tier: 2},
					{url: *testutils.MustParseUrl("http://127.0.0.1:4456"), tier: 2},
				},
				indexToDeprioritize: 4,
			},
			want: []Tracker{
				{url: *testutils.MustParseUrl("http://localhost:8080"), tier: 1},
				{url: *testutils.MustParseUrl("http://localhost:6549"), tier: 1},
				{url: *testutils.MustParseUrl("http://localhost:1236"), tier: 1},
				{url: *testutils.MustParseUrl("http://127.0.0.1:8080"), tier: 2},
				{url: *testutils.MustParseUrl("http://127.0.0.1:4456"), tier: 2},
			},
		},
		{
			name: "should not fail if tracker is alrady last of his tier",
			args: args{
				trackers: []Tracker{
					{url: *testutils.MustParseUrl("http://localhost:8080"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:1236"), tier: 1},
					{url: *testutils.MustParseUrl("http://127.0.0.1:8080"), tier: 2},
					{url: *testutils.MustParseUrl("http://127.0.0.1:4456"), tier: 2},
				},
				indexToDeprioritize: 1,
			},
			want: []Tracker{
				{url: *testutils.MustParseUrl("http://localhost:8080"), tier: 1},
				{url: *testutils.MustParseUrl("http://localhost:1236"), tier: 1},
				{url: *testutils.MustParseUrl("http://127.0.0.1:8080"), tier: 2},
				{url: *testutils.MustParseUrl("http://127.0.0.1:4456"), tier: 2},
			},
		},
		{
			name: "should not fail if tier has only one tracker",
			args: args{
				trackers: []Tracker{
					{url: *testutils.MustParseUrl("http://localhost:8080"), tier: 1},
					{url: *testutils.MustParseUrl("http://127.0.0.1:8080"), tier: 2},
					{url: *testutils.MustParseUrl("http://127.0.0.1:4456"), tier: 2},
				},
				indexToDeprioritize: 0,
			},
			want: []Tracker{
				{url: *testutils.MustParseUrl("http://localhost:8080"), tier: 1},
				{url: *testutils.MustParseUrl("http://127.0.0.1:8080"), tier: 2},
				{url: *testutils.MustParseUrl("http://127.0.0.1:4456"), tier: 2},
			},
		},
		{
			name: "should not fail if indexToDeprioritize is out of bound",
			args: args{
				trackers: []Tracker{
					{url: *testutils.MustParseUrl("http://localhost:8080"), tier: 1},
					{url: *testutils.MustParseUrl("http://127.0.0.1:8080"), tier: 2},
					{url: *testutils.MustParseUrl("http://127.0.0.1:4456"), tier: 2},
				},
				indexToDeprioritize: 10,
			},
			want: []Tracker{
				{url: *testutils.MustParseUrl("http://localhost:8080"), tier: 1},
				{url: *testutils.MustParseUrl("http://127.0.0.1:8080"), tier: 2},
				{url: *testutils.MustParseUrl("http://127.0.0.1:4456"), tier: 2},
			},
		},
		{
			name: "should not fail if there is only one tier with one tracker",
			args: args{
				trackers: []Tracker{
					{url: *testutils.MustParseUrl("http://localhost:8080"), tier: 1},
				},
				indexToDeprioritize: 0,
			},
			want: []Tracker{
				{url: *testutils.MustParseUrl("http://localhost:8080"), tier: 1},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deprioritizeTracker(tt.args.trackers, tt.args.indexToDeprioritize)
			assert.Equalf(t, tt.want, tt.args.trackers, "deprioritizeTracker(%v, %v)", tt.args.trackers, tt.args.indexToDeprioritize)
		})
	}
}

func Test_findTrackersInUse(t *testing.T) {
	type args struct {
		trackerList                 []Tracker
		announceToAllTiers          bool
		announceToAllTrackersInTier bool
	}
	tests := []struct {
		name string
		args args
		want []Tracker
	}{
		{
			name: "announceToAllTier=true && announceToAllTrackersInTier=true should return all trackers",
			args: args{
				trackerList: []Tracker{
					{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:11"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:20"), tier: 2},
					{url: *testutils.MustParseUrl("http://localhost:21"), tier: 2},
				},
				announceToAllTiers:          true,
				announceToAllTrackersInTier: true,
			},
			want: []Tracker{
				{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1},
				{url: *testutils.MustParseUrl("http://localhost:11"), tier: 1},
				{url: *testutils.MustParseUrl("http://localhost:20"), tier: 2},
				{url: *testutils.MustParseUrl("http://localhost:21"), tier: 2},
			},
		},
		{
			name: "announceToAllTier=true && announceToAllTrackersInTier=true should return all (non-disabled) trackers",
			args: args{
				trackerList: []Tracker{
					{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:11"), tier: 1, disabled: TrackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:20"), tier: 2, disabled: TrackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:21"), tier: 2},
				},
				announceToAllTiers:          true,
				announceToAllTrackersInTier: true,
			},
			want: []Tracker{
				{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1},
				{url: *testutils.MustParseUrl("http://localhost:21"), tier: 2},
			},
		},
		{
			name: "announceToAllTier=true && announceToAllTrackersInTier=true should return empty with empty list",
			args: args{
				trackerList:                 []Tracker{},
				announceToAllTiers:          true,
				announceToAllTrackersInTier: true,
			},
			want: []Tracker(nil),
		},
		{
			name: "announceToAllTier=true && announceToAllTrackersInTier=true should return empty list if all tracker are disabled",
			args: args{
				trackerList: []Tracker{
					{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1, disabled: TrackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:11"), tier: 1, disabled: TrackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:20"), tier: 2, disabled: TrackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:21"), tier: 2, disabled: TrackerDisabled{disabled: true}},
				},
				announceToAllTiers:          true,
				announceToAllTrackersInTier: true,
			},
			want: []Tracker(nil),
		},
		{
			name: "announceToAllTier=true && announceToAllTrackersInTier=false should return first tracker of each tier",
			args: args{
				trackerList: []Tracker{
					{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:11"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:20"), tier: 2},
					{url: *testutils.MustParseUrl("http://localhost:21"), tier: 2},
				},
				announceToAllTiers:          true,
				announceToAllTrackersInTier: false,
			},
			want: []Tracker{
				{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1},
				{url: *testutils.MustParseUrl("http://localhost:20"), tier: 2},
			},
		},
		{
			name: "announceToAllTier=true && announceToAllTrackersInTier=false should return first (non-disabled) tracker of each tier",
			args: args{
				trackerList: []Tracker{
					{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1, disabled: TrackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:11"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:20"), tier: 2, disabled: TrackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:21"), tier: 2},
				},
				announceToAllTiers:          true,
				announceToAllTrackersInTier: false,
			},
			want: []Tracker{
				{url: *testutils.MustParseUrl("http://localhost:11"), tier: 1},
				{url: *testutils.MustParseUrl("http://localhost:21"), tier: 2},
			},
		},
		{
			name: "announceToAllTier=true && announceToAllTrackersInTier=false should return empty with empty list",
			args: args{
				trackerList:                 []Tracker{},
				announceToAllTiers:          true,
				announceToAllTrackersInTier: false,
			},
			want: []Tracker(nil),
		},
		{
			name: "announceToAllTier=true && announceToAllTrackersInTier=false should return empty list if all tracker are disabled",
			args: args{
				trackerList: []Tracker{
					{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1, disabled: TrackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:11"), tier: 1, disabled: TrackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:20"), tier: 2, disabled: TrackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:21"), tier: 2, disabled: TrackerDisabled{disabled: true}},
				},
				announceToAllTiers:          true,
				announceToAllTrackersInTier: false,
			},
			want: []Tracker(nil),
		},
		{
			name: "announceToAllTier=false && announceToAllTrackersInTier=true should return all trackers of the first tier",
			args: args{
				trackerList: []Tracker{
					{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:11"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:20"), tier: 2},
					{url: *testutils.MustParseUrl("http://localhost:21"), tier: 2},
				},
				announceToAllTiers:          false,
				announceToAllTrackersInTier: true,
			},
			want: []Tracker{
				{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1},
				{url: *testutils.MustParseUrl("http://localhost:11"), tier: 1},
			},
		},
		{
			name: "announceToAllTier=false && announceToAllTrackersInTier=true should return all (non-disabled) trackers of the first tier that contains a (non-disabled) tracker",
			args: args{
				trackerList: []Tracker{
					{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1, disabled: TrackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:11"), tier: 1, disabled: TrackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:20"), tier: 2},
					{url: *testutils.MustParseUrl("http://localhost:21"), tier: 2, disabled: TrackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:22"), tier: 2},
				},
				announceToAllTiers:          false,
				announceToAllTrackersInTier: true,
			},
			want: []Tracker{
				{url: *testutils.MustParseUrl("http://localhost:20"), tier: 2},
				{url: *testutils.MustParseUrl("http://localhost:22"), tier: 2},
			},
		},
		{
			name: "announceToAllTier=false && announceToAllTrackersInTier=true should return empty with empty list",
			args: args{
				trackerList:                 []Tracker{},
				announceToAllTiers:          false,
				announceToAllTrackersInTier: true,
			},
			want: []Tracker(nil),
		},
		{
			name: "announceToAllTier=false && announceToAllTrackersInTier=true should return empty list if all tracker are disabled",
			args: args{
				trackerList: []Tracker{
					{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1, disabled: TrackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:11"), tier: 1, disabled: TrackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:20"), tier: 2, disabled: TrackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:21"), tier: 2, disabled: TrackerDisabled{disabled: true}},
				},
				announceToAllTiers:          false,
				announceToAllTrackersInTier: true,
			},
			want: []Tracker(nil),
		},
		{
			name: "announceToAllTier=false && announceToAllTrackersInTier=false should return first tracker",
			args: args{
				trackerList: []Tracker{
					{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:11"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:20"), tier: 2},
					{url: *testutils.MustParseUrl("http://localhost:21"), tier: 2},
				},
				announceToAllTiers:          false,
				announceToAllTrackersInTier: false,
			},
			want: []Tracker{
				{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1},
			},
		},
		{
			name: "announceToAllTier=false && announceToAllTrackersInTier=false should return first (non-disabled) tracker",
			args: args{
				trackerList: []Tracker{
					{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1, disabled: TrackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:11"), tier: 1, disabled: TrackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:20"), tier: 2, disabled: TrackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:21"), tier: 2},
				},
				announceToAllTiers:          false,
				announceToAllTrackersInTier: false,
			},
			want: []Tracker{
				{url: *testutils.MustParseUrl("http://localhost:21"), tier: 2},
			},
		},
		{
			name: "announceToAllTier=false && announceToAllTrackersInTier=false should return empty with empty list",
			args: args{
				trackerList:                 []Tracker{},
				announceToAllTiers:          false,
				announceToAllTrackersInTier: false,
			},
			want: []Tracker(nil),
		},
		{
			name: "announceToAllTier=false && announceToAllTrackersInTier=false should return empty list if all tracker are disabled",
			args: args{
				trackerList: []Tracker{
					{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1, disabled: TrackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:11"), tier: 1, disabled: TrackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:20"), tier: 2, disabled: TrackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:21"), tier: 2, disabled: TrackerDisabled{disabled: true}},
				},
				announceToAllTiers:          false,
				announceToAllTrackersInTier: false,
			},
			want: []Tracker(nil),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, findTrackersInUse(tt.args.trackerList, tt.args.announceToAllTiers, tt.args.announceToAllTrackersInTier), "findTrackersInUse(%v, %v, %v)", tt.args.trackerList, tt.args.announceToAllTiers, tt.args.announceToAllTrackersInTier)
		})
	}
}

func TestTrackers_Succeed(t *testing.T) {
	trs := Trackers{
		trackers: []Tracker{
			{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1, announcesHistory: []AnnounceHistory{}},
			{url: *testutils.MustParseUrl("http://localhost:11"), tier: 1, announcesHistory: []AnnounceHistory{}},
			{url: *testutils.MustParseUrl("http://localhost:12"), tier: 1, announcesHistory: []AnnounceHistory{}},
		},
		announceToAllTiers:          false,
		announceToAllTrackersInTier: false,
	}

	trs.Succeed(*testutils.MustParseUrl("http://localhost:10"), common.AnnounceResponse{
		Request:  common.AnnounceRequest{},
		Interval: 1 * time.Second,
		Leechers: 2,
		Seeders:  3,
	})

	assert.Equal(t, 1*time.Second, trs.trackers[0].announcesHistory[0].interval)
	assert.EqualValues(t, 2, trs.trackers[0].announcesHistory[0].leechers)
	assert.EqualValues(t, 3, trs.trackers[0].announcesHistory[0].seeders)
}

func TestTrackers_Failed(t *testing.T) {
	trs := Trackers{
		trackers: []Tracker{
			{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1, announcesHistory: []AnnounceHistory{}},
			{url: *testutils.MustParseUrl("http://localhost:11"), tier: 1, announcesHistory: []AnnounceHistory{}},
			{url: *testutils.MustParseUrl("http://localhost:12"), tier: 1, announcesHistory: []AnnounceHistory{}},
		},
		announceToAllTiers:          false,
		announceToAllTrackersInTier: false,
	}

	trs.Failed(*testutils.MustParseUrl("http://localhost:10"), common.AnnounceResponseError{
		Request: common.AnnounceRequest{},
		Error:   fmt.Errorf("nop"),
	})

	// tacker should have been deprioritized (that's why index 2)
	assert.EqualValues(t, "http://localhost:10", trs.trackers[2].url.String())
	assert.EqualValues(t, "nop", trs.trackers[2].announcesHistory[0].error)
}
