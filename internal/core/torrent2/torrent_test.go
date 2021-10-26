package torrent2

import (
	"net/url"
	"strings"
	"testing"
	"time"
)

func Test_findAnnounceReadyTrackers_AnnounceToAllTiers_AnnounceToAllTrackers(t *testing.T) {
	announceToAllTier := true
	announceToAllTracker := true
	type args struct {
		trackers             []*trackerImpl
		announceToAllTier    bool
		announceToAllTracker bool
	}
	tests := []struct {
		name           string
		args           args
		wantTrackerUrl []string
	}{
		{name: "should-return-all-ready", args: args{announceToAllTier: announceToAllTier, announceToAllTracker: announceToAllTracker, trackers: []*trackerImpl{
			&trackerImpl{url: &url.URL{Path: "0-0"}, tier: 0, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "0-1"}, tier: 0, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-0"}, tier: 1, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-1"}, tier: 1, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
		}}, wantTrackerUrl: []string{"0-0", "0-1", "1-0", "1-1"}},
		{name: "should-not-filter-not-working", args: args{announceToAllTier: announceToAllTier, announceToAllTracker: announceToAllTracker, trackers: []*trackerImpl{
			&trackerImpl{url: &url.URL{Path: "0-0"}, tier: 0, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 3, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "0-1"}, tier: 0, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-0"}, tier: 1, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 3, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-1"}, tier: 1, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
		}}, wantTrackerUrl: []string{"0-0", "0-1", "1-0", "1-1"}},
		{name: "should-filter-future-nextAnnounce", args: args{announceToAllTier: announceToAllTier, announceToAllTracker: announceToAllTracker, trackers: []*trackerImpl{
			&trackerImpl{url: &url.URL{Path: "0-0"}, tier: 0, enabled: true, state: &trackerState{nextAnnounce: time.Now().Add(1 * time.Hour), fails: 3, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "0-1"}, tier: 0, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-0"}, tier: 1, enabled: true, state: &trackerState{nextAnnounce: time.Now().Add(1 * time.Hour), fails: 3, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-1"}, tier: 1, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
		}}, wantTrackerUrl: []string{"0-1", "1-1"}},
		{name: "should-filter-disabled", args: args{announceToAllTier: announceToAllTier, announceToAllTracker: announceToAllTracker, trackers: []*trackerImpl{
			&trackerImpl{url: &url.URL{Path: "0-0"}, tier: 0, enabled: false, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "0-1"}, tier: 0, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-0"}, tier: 1, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-1"}, tier: 1, enabled: false, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
		}}, wantTrackerUrl: []string{"0-1", "1-0"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findAnnounceReadyTrackers(tt.args.trackers, tt.args.announceToAllTier, tt.args.announceToAllTracker); !ensureUrlInOrder(got, tt.wantTrackerUrl) {
				t.Errorf("findAnnounceReadyTrackers() = %v, want %v", prettyPrintTrackersUrl(got), prettyPrintUrl(tt.wantTrackerUrl))
			}
		})
	}
}

func Test_findAnnounceReadyTrackers_AnnounceToAllTiers_NotAnnounceToAllTrackers(t *testing.T) {
	announceToAllTier := true
	announceToAllTracker := false
	type args struct {
		trackers             []*trackerImpl
		announceToAllTier    bool
		announceToAllTracker bool
	}
	tests := []struct {
		name           string
		args           args
		wantTrackerUrl []string
	}{
		{name: "should-return-only-first-in-tier", args: args{announceToAllTier: announceToAllTier, announceToAllTracker: announceToAllTracker, trackers: []*trackerImpl{
			&trackerImpl{url: &url.URL{Path: "0-0"}, tier: 0, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "0-1"}, tier: 0, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-0"}, tier: 1, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-1"}, tier: 1, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
		}}, wantTrackerUrl: []string{"0-0", "1-0"}},
		{name: "should-return-all-ready-even-if-not-working", args: args{announceToAllTier: announceToAllTier, announceToAllTracker: announceToAllTracker, trackers: []*trackerImpl{
			&trackerImpl{url: &url.URL{Path: "0-0"}, tier: 0, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 3, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "0-1"}, tier: 0, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-0"}, tier: 1, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 3, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-1"}, tier: 1, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
		}}, wantTrackerUrl: []string{"0-0", "1-0"}},
		{name: "should-not-include-fallback-if-main-not-ready-but-working", args: args{announceToAllTier: announceToAllTier, announceToAllTracker: announceToAllTracker, trackers: []*trackerImpl{
			&trackerImpl{url: &url.URL{Path: "0-0"}, tier: 0, enabled: true, state: &trackerState{nextAnnounce: time.Now().Add(1 * time.Hour), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "0-1"}, tier: 0, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-0"}, tier: 1, enabled: true, state: &trackerState{nextAnnounce: time.Now().Add(1 * time.Hour), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-1"}, tier: 1, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
		}}, wantTrackerUrl: []string{}},
		{name: "should-filter-disabled", args: args{announceToAllTier: announceToAllTier, announceToAllTracker: announceToAllTracker, trackers: []*trackerImpl{
			&trackerImpl{url: &url.URL{Path: "0-0"}, tier: 0, enabled: false, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "0-1"}, tier: 0, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-0"}, tier: 1, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-1"}, tier: 1, enabled: false, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
		}}, wantTrackerUrl: []string{"0-1", "1-0"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findAnnounceReadyTrackers(tt.args.trackers, tt.args.announceToAllTier, tt.args.announceToAllTracker); !ensureUrlInOrder(got, tt.wantTrackerUrl) {
				t.Errorf("findAnnounceReadyTrackers() = %v, want %v", prettyPrintTrackersUrl(got), prettyPrintUrl(tt.wantTrackerUrl))
			}
		})
	}
}

func Test_findAnnounceReadyTrackers_NotAnnounceToAllTiers_AnnounceToAllTrackers(t *testing.T) {
	announceToAllTier := false
	announceToAllTracker := true
	type args struct {
		trackers             []*trackerImpl
		announceToAllTier    bool
		announceToAllTracker bool
	}
	tests := []struct {
		name           string
		args           args
		wantTrackerUrl []string
	}{
		{name: "should-return-only-first-tier", args: args{announceToAllTier: announceToAllTier, announceToAllTracker: announceToAllTracker, trackers: []*trackerImpl{
			&trackerImpl{url: &url.URL{Path: "0-0"}, tier: 0, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "0-1"}, tier: 0, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-0"}, tier: 1, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-1"}, tier: 1, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
		}}, wantTrackerUrl: []string{"0-0", "0-1"}},
		{name: "should-return-second-tier-if-none-of-first-tier-are-ready-and-not-working", args: args{announceToAllTier: announceToAllTier, announceToAllTracker: announceToAllTracker, trackers: []*trackerImpl{
			&trackerImpl{url: &url.URL{Path: "0-0"}, tier: 0, enabled: true, state: &trackerState{nextAnnounce: time.Now().Add(1 * time.Hour), fails: 5, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "0-1"}, tier: 0, enabled: true, state: &trackerState{nextAnnounce: time.Now().Add(1 * time.Hour), fails: 5, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-0"}, tier: 1, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-1"}, tier: 1, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
		}}, wantTrackerUrl: []string{"1-0", "1-1"}},
		{name: "should-return-all-ready-even-if-not-working", args: args{announceToAllTier: announceToAllTier, announceToAllTracker: announceToAllTracker, trackers: []*trackerImpl{
			&trackerImpl{url: &url.URL{Path: "0-0"}, tier: 0, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 3, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "0-1"}, tier: 0, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 3, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-0"}, tier: 1, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-1"}, tier: 1, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
		}}, wantTrackerUrl: []string{"0-0", "0-1"}},
		{name: "should-include-fallback-tracker-if-main-not-ready-but-working", args: args{announceToAllTier: announceToAllTier, announceToAllTracker: announceToAllTracker, trackers: []*trackerImpl{
			&trackerImpl{url: &url.URL{Path: "0-0"}, tier: 0, enabled: true, state: &trackerState{nextAnnounce: time.Now().Add(1 * time.Hour), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "0-1"}, tier: 0, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-0"}, tier: 1, enabled: true, state: &trackerState{nextAnnounce: time.Now().Add(1 * time.Hour), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-1"}, tier: 1, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
		}}, wantTrackerUrl: []string{"0-1"}},
		{name: "should-filter-disabled", args: args{announceToAllTier: announceToAllTier, announceToAllTracker: announceToAllTracker, trackers: []*trackerImpl{
			&trackerImpl{url: &url.URL{Path: "0-0"}, tier: 0, enabled: false, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "0-1"}, tier: 0, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-0"}, tier: 1, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-1"}, tier: 1, enabled: false, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
		}}, wantTrackerUrl: []string{"0-1"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findAnnounceReadyTrackers(tt.args.trackers, tt.args.announceToAllTier, tt.args.announceToAllTracker); !ensureUrlInOrder(got, tt.wantTrackerUrl) {
				t.Errorf("findAnnounceReadyTrackers() = %v, want %v", prettyPrintTrackersUrl(got), prettyPrintUrl(tt.wantTrackerUrl))
			}
		})
	}
}

func Test_findAnnounceReadyTrackers_NotAnnounceToAllTiers_NotAnnounceToAllTrackers(t *testing.T) {
	announceToAllTier := false
	announceToAllTracker := false
	type args struct {
		trackers             []*trackerImpl
		announceToAllTier    bool
		announceToAllTracker bool
	}
	tests := []struct {
		name           string
		args           args
		wantTrackerUrl []string
	}{
		{name: "should-return-only-first-tracker-in-first-tier", args: args{announceToAllTier: announceToAllTier, announceToAllTracker: announceToAllTracker, trackers: []*trackerImpl{
			&trackerImpl{url: &url.URL{Path: "0-0"}, tier: 0, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "0-1"}, tier: 0, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-0"}, tier: 1, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-1"}, tier: 1, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
		}}, wantTrackerUrl: []string{"0-0"}},
		{name: "should-return-second-tier-if-none-of-first-tier-are-ready-and-not-working", args: args{announceToAllTier: announceToAllTier, announceToAllTracker: announceToAllTracker, trackers: []*trackerImpl{
			&trackerImpl{url: &url.URL{Path: "0-0"}, tier: 0, enabled: true, state: &trackerState{nextAnnounce: time.Now().Add(1 * time.Hour), fails: 5, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "0-1"}, tier: 0, enabled: true, state: &trackerState{nextAnnounce: time.Now().Add(1 * time.Hour), fails: 5, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-0"}, tier: 1, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-1"}, tier: 1, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
		}}, wantTrackerUrl: []string{"1-0"}},
		{name: "should-return-first-tracker-in-first-tier-if-not-working-but-ready", args: args{announceToAllTier: announceToAllTier, announceToAllTracker: announceToAllTracker, trackers: []*trackerImpl{
			&trackerImpl{url: &url.URL{Path: "0-0"}, tier: 0, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 3, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "0-1"}, tier: 0, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-0"}, tier: 1, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 3, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-1"}, tier: 1, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
		}}, wantTrackerUrl: []string{"0-0"}},
		{name: "should-return-empty-if-main-not-ready-but-working", args: args{announceToAllTier: announceToAllTier, announceToAllTracker: announceToAllTracker, trackers: []*trackerImpl{
			&trackerImpl{url: &url.URL{Path: "0-0"}, tier: 0, enabled: true, state: &trackerState{nextAnnounce: time.Now().Add(1 * time.Hour), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "0-1"}, tier: 0, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-0"}, tier: 1, enabled: true, state: &trackerState{nextAnnounce: time.Now().Add(1 * time.Hour), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-1"}, tier: 1, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
		}}, wantTrackerUrl: []string{}},
		{name: "should-filter-disabled", args: args{announceToAllTier: announceToAllTier, announceToAllTracker: announceToAllTracker, trackers: []*trackerImpl{
			&trackerImpl{url: &url.URL{Path: "0-0"}, tier: 0, enabled: false, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "0-1"}, tier: 0, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-0"}, tier: 1, enabled: true, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
			&trackerImpl{url: &url.URL{Path: "1-1"}, tier: 1, enabled: false, state: &trackerState{nextAnnounce: time.Now(), fails: 0, startSent: false, updating: false}},
		}}, wantTrackerUrl: []string{"0-1"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findAnnounceReadyTrackers(tt.args.trackers, tt.args.announceToAllTier, tt.args.announceToAllTracker); !ensureUrlInOrder(got, tt.wantTrackerUrl) {
				t.Errorf("findAnnounceReadyTrackers() = %v, want %v", prettyPrintTrackersUrl(got), prettyPrintUrl(tt.wantTrackerUrl))
			}
		})
	}
}

func ensureUrlInOrder(trackers []*trackerImpl, urls []string) bool {
	if len(urls) == 0 && len(trackers) != 0 {
		return false
	}
	if len(urls) != len(trackers) {
		return false
	}
	for i := range urls {
		if trackers[i].url.Path != urls[i] {
			return false
		}
	}
	return true
}

func prettyPrintTrackersUrl(trackers []*trackerImpl) string {
	var b strings.Builder
	b.WriteString("{ ")
	for _, t := range trackers {
		b.WriteString(t.url.Path + " ")
	}
	b.WriteString("}")
	return b.String()
}

func prettyPrintUrl(urls []string) string {
	var b strings.Builder
	b.WriteString("{ ")
	for _, u := range urls {
		b.WriteString(u + " ")
	}
	b.WriteString("}")
	return b.String()
}
