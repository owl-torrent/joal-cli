package torrent

import (
	"fmt"
	"github.com/anthonyraymond/joal-cli/internal/utils/testutils"
	"github.com/stretchr/testify/assert"
	"net/url"
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

	announceList := [][]url.URL{
		{*testutils.MustParseUrl("udp://localhost:20/tier1/tr1"), *testutils.MustParseUrl("http://localhost:20/tier1/tr2")},
		{*testutils.MustParseUrl("http://localhost:20/tier2/tr1"), *testutils.MustParseUrl("https://localhost:20/tier2/tr2"), *testutils.MustParseUrl("udp://localhost:20/tier2/tr3")},
	}

	trackers, err := createTrackers(*testutils.MustParseUrl("http://localhost:1111/announce"), announceList, announcePolicy)
	assert.NoError(t, err)

	for _, tr := range trackers.trackers {
		assert.NotEqual(t, "http://localhost:1111/announce", tr.url.String(), "when supportAnnounceList is true, the single announce url should not be added to tracker list")
		assert.False(t, tr.disabled.isDisabled(), "no trackers should be disabled")
	}
	assert.Len(t, trackers.trackers, 5)
}

func TestCreateTrackers_shouldDisableUdpTrackers(t *testing.T) {
	announcePolicy := testAnnouncePolicy{
		supportHttpAnnounce: true,
		supportUdpAnnounce:  false,
		supportAnnounceList: true,
	}

	announceList := [][]url.URL{
		{*testutils.MustParseUrl("udp://localhost:20/tier1/tr1"), *testutils.MustParseUrl("http://localhost:20/tier1/tr2")},
		{*testutils.MustParseUrl("http://localhost:20/tier2/tr1"), *testutils.MustParseUrl("https://localhost:20/tier2/tr2"), *testutils.MustParseUrl("UDP://localhost:20/tier2/tr3")},
	}

	trackers, err := createTrackers(*testutils.MustParseUrl("http://localhost:1111/announce"), announceList, announcePolicy)
	assert.NoError(t, err)

	disabledCount := 0
	for _, tr := range trackers.trackers {
		if strings.HasPrefix(strings.ToLower(tr.url.Scheme), "udp") {
			disabledCount++
			assert.True(t, tr.disabled.isDisabled())
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

	announceList := [][]url.URL{
		{*testutils.MustParseUrl("udp://localhost:20/tier1/tr1"), *testutils.MustParseUrl("http://localhost:20/tier1/tr2"), *testutils.MustParseUrl("HTTP://localhost:20/tier1/tr3")},
		{*testutils.MustParseUrl("http://localhost:20/tier2/tr1"), *testutils.MustParseUrl("https://localhost:20/tier2/tr2"), *testutils.MustParseUrl("udp://localhost:20/tier2/tr3")},
	}

	trackers, err := createTrackers(*testutils.MustParseUrl("http://localhost:1111/announce"), announceList, announcePolicy)
	assert.NoError(t, err)

	disabledCount := 0
	for _, tr := range trackers.trackers {
		if strings.HasPrefix(strings.ToLower(tr.url.Scheme), "http") {
			disabledCount++
			assert.True(t, tr.disabled.isDisabled())
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

	announceList := [][]url.URL{
		{*testutils.MustParseUrl("udp://localhost:20/tier1/tr1"), *testutils.MustParseUrl("dop://localhost:20/tier1/tr2")},
		{*testutils.MustParseUrl("http://localhost:20/tier2/tr1"), *testutils.MustParseUrl("pat://localhost:20/tier2/tr2"), *testutils.MustParseUrl("udp://localhost:20/tier2/tr3")},
	}

	trackers, err := createTrackers(*testutils.MustParseUrl("http://localhost:1111/announce"), announceList, announcePolicy)
	assert.NoError(t, err)

	disabledCount := 0
	for _, tr := range trackers.trackers {
		if strings.HasPrefix(strings.ToLower(tr.url.Scheme), "dop") || strings.HasPrefix(strings.ToLower(tr.url.Scheme), "pat") {
			disabledCount++
			assert.True(t, tr.disabled.isDisabled())
			assert.Equal(t, announceProtocolNotSupported, tr.disabled)
		}
	}
	assert.Equal(t, 2, disabledCount)
}

func TestCreateTrackers_shouldDisableAllTrackerIfAnnounceListIsNotSupportedAndAddTheSingleAnnounceAsTier0(t *testing.T) {
	announcePolicy := testAnnouncePolicy{
		supportHttpAnnounce: true,
		supportUdpAnnounce:  true,
		supportAnnounceList: false,
	}

	announceList := [][]url.URL{
		{*testutils.MustParseUrl("udp://localhost:20/tier1/tr1"), *testutils.MustParseUrl("http://localhost:20/tier1/tr2")},
		{*testutils.MustParseUrl("http://localhost:20/tier2/tr1"), *testutils.MustParseUrl("https://localhost:20/tier2/tr2"), *testutils.MustParseUrl("udp://localhost:20/tier2/tr3")},
	}

	trackers, err := createTrackers(*testutils.MustParseUrl("http://localhost:1111/announce"), announceList, announcePolicy)
	assert.NoError(t, err)

	// When supportAnnounceList is false, the single "announce" tracker should be moved on alone on tier 1 and all tiers of announceList must be incremented by one
	assert.Len(t, trackers.trackers, 6, "there should be 6 trackers: 1 single announce + 5 of announce list")
	var trackersOfTierZero []tracker
	for _, tr := range trackers.trackers {
		if tr.tier == 0 {
			trackersOfTierZero = append(trackersOfTierZero, tr)
		}
	}
	assert.Len(t, trackersOfTierZero, 1, "getTier zero should contains only the single announce")
	assert.Equal(t, "http://localhost:1111/announce", trackersOfTierZero[0].url.String())

	// All trackers from announceList should be disabled
	disabledCount := 0
	for _, tr := range trackers.trackers {
		if tr.url.String() != "http://localhost:1111/announce" {
			assert.True(t, tr.disabled.isDisabled())
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

	announceList := [][]url.URL{
		{*testutils.MustParseUrl("udp://localhost:20/tier1/tr1"), *testutils.MustParseUrl("http://localhost:1111/announce")},
		{*testutils.MustParseUrl("http://localhost:20/tier2/tr1"), *testutils.MustParseUrl("http://localhost:1111/announce"), *testutils.MustParseUrl("udp://localhost:20/tier2/tr3")},
	}

	trackers, err := createTrackers(*testutils.MustParseUrl("http://localhost:1111/announce"), announceList, announcePolicy)
	assert.NoError(t, err)

	numberOfTimeFound := 0
	for _, tr := range trackers.trackers {
		if tr.url.String() == "http://localhost:1111/announce" {
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

	announceList := [][]url.URL{{}, {}}

	trackers, err := createTrackers(*testutils.MustParseUrl("http://localhost:1111/announce"), announceList, announcePolicy)
	assert.NoError(t, err)

	assert.Len(t, trackers.trackers, 1)
	assert.Equal(t, trackers.trackers[0].url.String(), "http://localhost:1111/announce")
}

func Test_findTrackerForUrl(t *testing.T) {
	trackers := []tracker{
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
	trackers := []tracker{
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
	trackers := []tracker{
		{url: *testutils.MustParseUrl("http://localhost:1234/announce")},
		{url: *testutils.MustParseUrl("http://localhost:5678/announce")},
	}

	_, err := findTrackerForUrl(trackers, *testutils.MustParseUrl("http://localhost:9876/announce"))
	assert.Error(t, err)
}

func Test_deprioritizeTracker(t *testing.T) {
	type args struct {
		trackers            []tracker
		indexToDeprioritize int
	}
	tests := []struct {
		name string
		args args
		want []tracker
	}{
		{
			name: "should move tracker 0 last of his tier",
			args: args{
				trackers: []tracker{
					{url: *testutils.MustParseUrl("http://localhost:8080"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:1236"), tier: 1},
					{url: *testutils.MustParseUrl("http://127.0.0.1:8080"), tier: 2},
					{url: *testutils.MustParseUrl("http://127.0.0.1:4456"), tier: 2},
				},
				indexToDeprioritize: 0,
			},
			want: []tracker{
				{url: *testutils.MustParseUrl("http://localhost:1236"), tier: 1},
				{url: *testutils.MustParseUrl("http://localhost:8080"), tier: 1},
				{url: *testutils.MustParseUrl("http://127.0.0.1:8080"), tier: 2},
				{url: *testutils.MustParseUrl("http://127.0.0.1:4456"), tier: 2},
			},
		},
		{
			name: "should move tracker 1 last of his tier",
			args: args{
				trackers: []tracker{
					{url: *testutils.MustParseUrl("http://localhost:8080"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:6549"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:1236"), tier: 1},
					{url: *testutils.MustParseUrl("http://127.0.0.1:8080"), tier: 2},
					{url: *testutils.MustParseUrl("http://127.0.0.1:4456"), tier: 2},
				},
				indexToDeprioritize: 1,
			},
			want: []tracker{
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
				trackers: []tracker{
					{url: *testutils.MustParseUrl("http://localhost:8080"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:6549"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:1236"), tier: 1},
					{url: *testutils.MustParseUrl("http://127.0.0.1:8080"), tier: 2},
					{url: *testutils.MustParseUrl("http://127.0.0.1:4456"), tier: 2},
				},
				indexToDeprioritize: 3,
			},
			want: []tracker{
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
				trackers: []tracker{
					{url: *testutils.MustParseUrl("http://localhost:8080"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:6549"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:1236"), tier: 1},
					{url: *testutils.MustParseUrl("http://127.0.0.1:8080"), tier: 2},
					{url: *testutils.MustParseUrl("http://127.0.0.1:4456"), tier: 2},
				},
				indexToDeprioritize: 4,
			},
			want: []tracker{
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
				trackers: []tracker{
					{url: *testutils.MustParseUrl("http://localhost:8080"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:1236"), tier: 1},
					{url: *testutils.MustParseUrl("http://127.0.0.1:8080"), tier: 2},
					{url: *testutils.MustParseUrl("http://127.0.0.1:4456"), tier: 2},
				},
				indexToDeprioritize: 1,
			},
			want: []tracker{
				{url: *testutils.MustParseUrl("http://localhost:8080"), tier: 1},
				{url: *testutils.MustParseUrl("http://localhost:1236"), tier: 1},
				{url: *testutils.MustParseUrl("http://127.0.0.1:8080"), tier: 2},
				{url: *testutils.MustParseUrl("http://127.0.0.1:4456"), tier: 2},
			},
		},
		{
			name: "should not fail if tier has only one tracker",
			args: args{
				trackers: []tracker{
					{url: *testutils.MustParseUrl("http://localhost:8080"), tier: 1},
					{url: *testutils.MustParseUrl("http://127.0.0.1:8080"), tier: 2},
					{url: *testutils.MustParseUrl("http://127.0.0.1:4456"), tier: 2},
				},
				indexToDeprioritize: 0,
			},
			want: []tracker{
				{url: *testutils.MustParseUrl("http://localhost:8080"), tier: 1},
				{url: *testutils.MustParseUrl("http://127.0.0.1:8080"), tier: 2},
				{url: *testutils.MustParseUrl("http://127.0.0.1:4456"), tier: 2},
			},
		},
		{
			name: "should not fail if indexToDeprioritize is out of bound",
			args: args{
				trackers: []tracker{
					{url: *testutils.MustParseUrl("http://localhost:8080"), tier: 1},
					{url: *testutils.MustParseUrl("http://127.0.0.1:8080"), tier: 2},
					{url: *testutils.MustParseUrl("http://127.0.0.1:4456"), tier: 2},
				},
				indexToDeprioritize: 10,
			},
			want: []tracker{
				{url: *testutils.MustParseUrl("http://localhost:8080"), tier: 1},
				{url: *testutils.MustParseUrl("http://127.0.0.1:8080"), tier: 2},
				{url: *testutils.MustParseUrl("http://127.0.0.1:4456"), tier: 2},
			},
		},
		{
			name: "should not fail if there is only one tier with one tracker",
			args: args{
				trackers: []tracker{
					{url: *testutils.MustParseUrl("http://localhost:8080"), tier: 1},
				},
				indexToDeprioritize: 0,
			},
			want: []tracker{
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
		trackerList                 []tracker
		announceToAllTiers          bool
		announceToAllTrackersInTier bool
	}
	tests := []struct {
		name string
		args args
		want []tracker
	}{
		{
			name: "announceToAllTier=true && announceToAllTrackersInTier=true should return all trackers",
			args: args{
				trackerList: []tracker{
					{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:11"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:20"), tier: 2},
					{url: *testutils.MustParseUrl("http://localhost:21"), tier: 2},
				},
				announceToAllTiers:          true,
				announceToAllTrackersInTier: true,
			},
			want: []tracker{
				{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1},
				{url: *testutils.MustParseUrl("http://localhost:11"), tier: 1},
				{url: *testutils.MustParseUrl("http://localhost:20"), tier: 2},
				{url: *testutils.MustParseUrl("http://localhost:21"), tier: 2},
			},
		},
		{
			name: "announceToAllTier=true && announceToAllTrackersInTier=true should return all (non-disabled) trackers",
			args: args{
				trackerList: []tracker{
					{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:11"), tier: 1, disabled: trackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:20"), tier: 2, disabled: trackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:21"), tier: 2},
				},
				announceToAllTiers:          true,
				announceToAllTrackersInTier: true,
			},
			want: []tracker{
				{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1},
				{url: *testutils.MustParseUrl("http://localhost:21"), tier: 2},
			},
		},
		{
			name: "announceToAllTier=true && announceToAllTrackersInTier=true should return empty with empty list",
			args: args{
				trackerList:                 []tracker{},
				announceToAllTiers:          true,
				announceToAllTrackersInTier: true,
			},
			want: []tracker(nil),
		},
		{
			name: "announceToAllTier=true && announceToAllTrackersInTier=true should return empty list if all tracker are disabled",
			args: args{
				trackerList: []tracker{
					{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1, disabled: trackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:11"), tier: 1, disabled: trackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:20"), tier: 2, disabled: trackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:21"), tier: 2, disabled: trackerDisabled{disabled: true}},
				},
				announceToAllTiers:          true,
				announceToAllTrackersInTier: true,
			},
			want: []tracker(nil),
		},
		{
			name: "announceToAllTier=true && announceToAllTrackersInTier=false should return first tracker of each tier",
			args: args{
				trackerList: []tracker{
					{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:11"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:20"), tier: 2},
					{url: *testutils.MustParseUrl("http://localhost:21"), tier: 2},
				},
				announceToAllTiers:          true,
				announceToAllTrackersInTier: false,
			},
			want: []tracker{
				{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1},
				{url: *testutils.MustParseUrl("http://localhost:20"), tier: 2},
			},
		},
		{
			name: "announceToAllTier=true && announceToAllTrackersInTier=false should return first (non-disabled) tracker of each tier",
			args: args{
				trackerList: []tracker{
					{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1, disabled: trackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:11"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:20"), tier: 2, disabled: trackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:21"), tier: 2},
				},
				announceToAllTiers:          true,
				announceToAllTrackersInTier: false,
			},
			want: []tracker{
				{url: *testutils.MustParseUrl("http://localhost:11"), tier: 1},
				{url: *testutils.MustParseUrl("http://localhost:21"), tier: 2},
			},
		},
		{
			name: "announceToAllTier=true && announceToAllTrackersInTier=false should return empty with empty list",
			args: args{
				trackerList:                 []tracker{},
				announceToAllTiers:          true,
				announceToAllTrackersInTier: false,
			},
			want: []tracker(nil),
		},
		{
			name: "announceToAllTier=true && announceToAllTrackersInTier=false should return empty list if all tracker are disabled",
			args: args{
				trackerList: []tracker{
					{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1, disabled: trackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:11"), tier: 1, disabled: trackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:20"), tier: 2, disabled: trackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:21"), tier: 2, disabled: trackerDisabled{disabled: true}},
				},
				announceToAllTiers:          true,
				announceToAllTrackersInTier: false,
			},
			want: []tracker(nil),
		},
		{
			name: "announceToAllTier=false && announceToAllTrackersInTier=true should return all trackers of the first tier",
			args: args{
				trackerList: []tracker{
					{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:11"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:20"), tier: 2},
					{url: *testutils.MustParseUrl("http://localhost:21"), tier: 2},
				},
				announceToAllTiers:          false,
				announceToAllTrackersInTier: true,
			},
			want: []tracker{
				{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1},
				{url: *testutils.MustParseUrl("http://localhost:11"), tier: 1},
			},
		},
		{
			name: "announceToAllTier=false && announceToAllTrackersInTier=true should return all (non-disabled) trackers of the first tier that contains a (non-disabled) tracker",
			args: args{
				trackerList: []tracker{
					{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1, disabled: trackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:11"), tier: 1, disabled: trackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:20"), tier: 2},
					{url: *testutils.MustParseUrl("http://localhost:21"), tier: 2, disabled: trackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:22"), tier: 2},
				},
				announceToAllTiers:          false,
				announceToAllTrackersInTier: true,
			},
			want: []tracker{
				{url: *testutils.MustParseUrl("http://localhost:20"), tier: 2},
				{url: *testutils.MustParseUrl("http://localhost:22"), tier: 2},
			},
		},
		{
			name: "announceToAllTier=false && announceToAllTrackersInTier=true should return empty with empty list",
			args: args{
				trackerList:                 []tracker{},
				announceToAllTiers:          false,
				announceToAllTrackersInTier: true,
			},
			want: []tracker(nil),
		},
		{
			name: "announceToAllTier=false && announceToAllTrackersInTier=true should return empty list if all tracker are disabled",
			args: args{
				trackerList: []tracker{
					{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1, disabled: trackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:11"), tier: 1, disabled: trackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:20"), tier: 2, disabled: trackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:21"), tier: 2, disabled: trackerDisabled{disabled: true}},
				},
				announceToAllTiers:          false,
				announceToAllTrackersInTier: true,
			},
			want: []tracker(nil),
		},
		{
			name: "announceToAllTier=false && announceToAllTrackersInTier=false should return first tracker",
			args: args{
				trackerList: []tracker{
					{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:11"), tier: 1},
					{url: *testutils.MustParseUrl("http://localhost:20"), tier: 2},
					{url: *testutils.MustParseUrl("http://localhost:21"), tier: 2},
				},
				announceToAllTiers:          false,
				announceToAllTrackersInTier: false,
			},
			want: []tracker{
				{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1},
			},
		},
		{
			name: "announceToAllTier=false && announceToAllTrackersInTier=false should return first (non-disabled) tracker",
			args: args{
				trackerList: []tracker{
					{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1, disabled: trackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:11"), tier: 1, disabled: trackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:20"), tier: 2, disabled: trackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:21"), tier: 2},
				},
				announceToAllTiers:          false,
				announceToAllTrackersInTier: false,
			},
			want: []tracker{
				{url: *testutils.MustParseUrl("http://localhost:21"), tier: 2},
			},
		},
		{
			name: "announceToAllTier=false && announceToAllTrackersInTier=false should return empty with empty list",
			args: args{
				trackerList:                 []tracker{},
				announceToAllTiers:          false,
				announceToAllTrackersInTier: false,
			},
			want: []tracker(nil),
		},
		{
			name: "announceToAllTier=false && announceToAllTrackersInTier=false should return empty list if all tracker are disabled",
			args: args{
				trackerList: []tracker{
					{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1, disabled: trackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:11"), tier: 1, disabled: trackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:20"), tier: 2, disabled: trackerDisabled{disabled: true}},
					{url: *testutils.MustParseUrl("http://localhost:21"), tier: 2, disabled: trackerDisabled{disabled: true}},
				},
				announceToAllTiers:          false,
				announceToAllTrackersInTier: false,
			},
			want: []tracker(nil),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, findTrackersInUse(tt.args.trackerList, tt.args.announceToAllTiers, tt.args.announceToAllTrackersInTier), "findTrackersInUse(%v, %v, %v)", tt.args.trackerList, tt.args.announceToAllTiers, tt.args.announceToAllTrackersInTier)
		})
	}
}

func TestTrackers_Succeed(t *testing.T) {
	trs := trackerPool{
		trackers: []tracker{
			{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1, announcesHistory: []announceHistory{}},
			{url: *testutils.MustParseUrl("http://localhost:11"), tier: 1, announcesHistory: []announceHistory{}},
			{url: *testutils.MustParseUrl("http://localhost:12"), tier: 1, announcesHistory: []announceHistory{}},
		},
		announceToAllTiers:          false,
		announceToAllTrackersInTier: false,
	}

	trs.succeed(*testutils.MustParseUrl("http://localhost:10"), TrackerAnnounceResponse{
		Request:  TrackerAnnounceRequest{},
		Interval: 1 * time.Second,
		Leechers: 2,
		Seeders:  3,
	})

	assert.Equal(t, 1*time.Second, trs.trackers[0].announcesHistory[0].interval)
	assert.EqualValues(t, 2, trs.trackers[0].announcesHistory[0].leechers)
	assert.EqualValues(t, 3, trs.trackers[0].announcesHistory[0].seeders)
}

func TestTrackers_Failed(t *testing.T) {
	trs := trackerPool{
		trackers: []tracker{
			{url: *testutils.MustParseUrl("http://localhost:10"), tier: 1, announcesHistory: []announceHistory{}},
			{url: *testutils.MustParseUrl("http://localhost:11"), tier: 1, announcesHistory: []announceHistory{}},
			{url: *testutils.MustParseUrl("http://localhost:12"), tier: 1, announcesHistory: []announceHistory{}},
		},
		announceToAllTiers:          false,
		announceToAllTrackersInTier: false,
	}

	trs.failed(*testutils.MustParseUrl("http://localhost:10"), TrackerAnnounceResponseError{
		Request: TrackerAnnounceRequest{},
		Error:   fmt.Errorf("nop"),
	})

	// tacker should have been deprioritized (that's why index 2)
	assert.EqualValues(t, "http://localhost:10", trs.trackers[2].url.String())
	assert.EqualValues(t, "nop", trs.trackers[2].announcesHistory[0].error)
}
