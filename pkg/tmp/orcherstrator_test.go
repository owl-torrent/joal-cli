package tmp

/*
import (
	"github.com/anthonyraymond/joal-cli/pkg/torrent"
	"net/url"
	"reflect"
	"sort"
	"testing"
)

func Test_newTierAnnouncer(t *testing.T) {
	type args struct {
		announceList                [][]string
		announceToAllTiers          bool
		announceToAllTrackersInTier bool
	}
	tests := []struct {
		name    string
		args    args
		want    TiersAnnouncer
		wantErr bool
	}{
		{name: "shouldFailWhenEmpty", args: args{announceList: [][]string{}, announceToAllTiers: true, announceToAllTrackersInTier: true}, want: nil, wantErr: true},
		{name: "shouldFailWhenEmpty", args: args{announceList: [][]string{{}}, announceToAllTiers: true, announceToAllTrackersInTier: true}, want: nil, wantErr: true},
		{name: "shouldFailWhenEmpty", args: args{announceList: [][]string{{}, {}}, announceToAllTiers: true, announceToAllTrackersInTier: true}, want: nil, wantErr: true},
		{name: "shouldCreateAllTierAllTracker", args: args{announceList: [][]string{{"http://localhost"}}, announceToAllTiers: true, announceToAllTrackersInTier: true},
			want: &AllTiersAnnouncer{
				tiers: []TrackersAnnouncer{
					&AllTrackersAnnouncer{trackers: []url.URL{*torrent.mustParseUrl("http://localhost")}},
				},
			},
			wantErr: false,
		},
		{name: "shouldCreateAllTierAllTracker", args: args{announceList: [][]string{{"http://localhost-t1-1", "http://localhost-t1-2"}}, announceToAllTiers: true, announceToAllTrackersInTier: true},
			want: &AllTiersAnnouncer{
				tiers: []TrackersAnnouncer{
					&AllTrackersAnnouncer{trackers: []url.URL{*torrent.mustParseUrl("http://localhost-t1-1"), *torrent.mustParseUrl("http://localhost-t1-2")}},
				},
			},
			wantErr: false,
		},
		{name: "shouldCreateAllTierAllTracker", args: args{announceList: [][]string{{"http://localhost-t1-1", "http://localhost-t1-2"}, {"http://localhost-t2-1", "http://localhost-t2-2"}}, announceToAllTiers: true, announceToAllTrackersInTier: true},
			want: &AllTiersAnnouncer{
				tiers: []TrackersAnnouncer{
					&AllTrackersAnnouncer{trackers: []url.URL{*torrent.mustParseUrl("http://localhost-t1-1"), *torrent.mustParseUrl("http://localhost-t1-2")}},
					&AllTrackersAnnouncer{trackers: []url.URL{*torrent.mustParseUrl("http://localhost-t2-1"), *torrent.mustParseUrl("http://localhost-t2-2")}},
				},
			},
			wantErr: false,
		},

		{name: "shouldCreateAllTierFallbackTracker", args: args{announceList: [][]string{{"http://localhost"}}, announceToAllTiers: true, announceToAllTrackersInTier: false},
			want: &AllTiersAnnouncer{
				tiers: []TrackersAnnouncer{
					&FallbackTrackersAnnouncer{trackers: []url.URL{*torrent.mustParseUrl("http://localhost")}},
				},
			},
			wantErr: false,
		},
		{name: "shouldCreateAllTierFallbackTracker", args: args{announceList: [][]string{{"http://localhost-t1-1", "http://localhost-t1-2"}}, announceToAllTiers: true, announceToAllTrackersInTier: false},
			want: &AllTiersAnnouncer{
				tiers: []TrackersAnnouncer{
					&FallbackTrackersAnnouncer{trackers: []url.URL{*torrent.mustParseUrl("http://localhost-t1-1"), *torrent.mustParseUrl("http://localhost-t1-2")}},
				},
			},
			wantErr: false,
		},
		{name: "shouldCreateAllTierFallbackTracker", args: args{announceList: [][]string{{"http://localhost-t1-1", "http://localhost-t1-2"}, {"http://localhost-t2-1", "http://localhost-t2-2"}}, announceToAllTiers: true, announceToAllTrackersInTier: false},
			want: &AllTiersAnnouncer{
				tiers: []TrackersAnnouncer{
					&FallbackTrackersAnnouncer{trackers: []url.URL{*torrent.mustParseUrl("http://localhost-t1-1"), *torrent.mustParseUrl("http://localhost-t1-2")}},
					&FallbackTrackersAnnouncer{trackers: []url.URL{*torrent.mustParseUrl("http://localhost-t2-1"), *torrent.mustParseUrl("http://localhost-t2-2")}},
				},
			},
			wantErr: false,
		},
		{name: "shouldCreateFallbackTierAllTracker", args: args{announceList: [][]string{{"http://localhost"}}, announceToAllTiers: false, announceToAllTrackersInTier: true},
			want: &FallbackTiersAnnouncer{
				tiers: []TrackersAnnouncer{
					&AllTrackersAnnouncer{trackers: []url.URL{*torrent.mustParseUrl("http://localhost")}},
				},
			},
			wantErr: false,
		},
		{name: "shouldCreateFallbackTierAllTracker", args: args{announceList: [][]string{{"http://localhost-t1-1", "http://localhost-t1-2"}}, announceToAllTiers: false, announceToAllTrackersInTier: true},
			want: &FallbackTiersAnnouncer{
				tiers: []TrackersAnnouncer{
					&AllTrackersAnnouncer{trackers: []url.URL{*torrent.mustParseUrl("http://localhost-t1-1"), *torrent.mustParseUrl("http://localhost-t1-2")}},
				},
			},
			wantErr: false,
		},
		{name: "shouldCreateFallbackTierAllTracker", args: args{announceList: [][]string{{"http://localhost-t1-1", "http://localhost-t1-2"}, {"http://localhost-t2-1", "http://localhost-t2-2"}}, announceToAllTiers: false, announceToAllTrackersInTier: true},
			want: &FallbackTiersAnnouncer{
				tiers: []TrackersAnnouncer{
					&AllTrackersAnnouncer{trackers: []url.URL{*torrent.mustParseUrl("http://localhost-t1-1"), *torrent.mustParseUrl("http://localhost-t1-2")}},
					&AllTrackersAnnouncer{trackers: []url.URL{*torrent.mustParseUrl("http://localhost-t2-1"), *torrent.mustParseUrl("http://localhost-t2-2")}},
				},
			},
			wantErr: false,
		},
		{name: "shouldCreateFallbackTierFallbackTracker", args: args{announceList: [][]string{{"http://localhost"}}, announceToAllTiers: false, announceToAllTrackersInTier: false},
			want: &FallbackTiersAnnouncer{
				tiers: []TrackersAnnouncer{
					&FallbackTrackersAnnouncer{trackers: []url.URL{*torrent.mustParseUrl("http://localhost")}},
				},
			},
			wantErr: false,
		},
		{name: "shouldCreateFallbackTierFallbackTracker", args: args{announceList: [][]string{{"http://localhost-t1-1", "http://localhost-t1-2"}}, announceToAllTiers: false, announceToAllTrackersInTier: false},
			want: &FallbackTiersAnnouncer{
				tiers: []TrackersAnnouncer{
					&FallbackTrackersAnnouncer{trackers: []url.URL{*torrent.mustParseUrl("http://localhost-t1-1"), *torrent.mustParseUrl("http://localhost-t1-2")}},
				},
			},
			wantErr: false,
		},
		{name: "shouldCreateFallbackTierFallbackTracker", args: args{announceList: [][]string{{"http://localhost-t1-1", "http://localhost-t1-2"}, {"http://localhost-t2-1", "http://localhost-t2-2"}}, announceToAllTiers: false, announceToAllTrackersInTier: false},
			want: &FallbackTiersAnnouncer{
				tiers: []TrackersAnnouncer{
					&FallbackTrackersAnnouncer{trackers: []url.URL{*torrent.mustParseUrl("http://localhost-t1-1"), *torrent.mustParseUrl("http://localhost-t1-2")}},
					&FallbackTrackersAnnouncer{trackers: []url.URL{*torrent.mustParseUrl("http://localhost-t2-1"), *torrent.mustParseUrl("http://localhost-t2-2")}},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newTierAnnouncer(nil, tt.args.announceList, tt.args.announceToAllTiers, tt.args.announceToAllTrackersInTier)
			if (err != nil) != tt.wantErr {
				t.Errorf("newTierAnnouncer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// the trackers url are shuffled, we sort them on every instances (got and want) to make the comparison order agnostic
			sortTrackerUris(got)
			sortTrackerUris(tt.want)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newTierAnnouncer() got = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func sortTrackerUris(i TiersAnnouncer) {
	if i == nil {
		return
	}
	if ta, ok := i.(*AllTiersAnnouncer); ok {
		if ta.tiers == nil {
			return
		}
		for _, tier := range ta.tiers {
			if trackers, ok := tier.(*AllTrackersAnnouncer); ok {
				sort.Slice(trackers.trackers, func(i, j int) bool { return trackers.trackers[i].String() < trackers.trackers[j].String() })
			}
			if trackers, ok := tier.(*FallbackTrackersAnnouncer); ok {
				sort.Slice(trackers.trackers, func(i, j int) bool { return trackers.trackers[i].String() < trackers.trackers[j].String() })
			}
		}
	}
	if ta, ok := i.(*FallbackTiersAnnouncer); ok {
		if ta.tiers == nil {
			return
		}
		for _, tier := range ta.tiers {
			if trackers, ok := tier.(*AllTrackersAnnouncer); ok {
				sort.Slice(trackers.trackers, func(i, j int) bool { return trackers.trackers[i].String() < trackers.trackers[j].String() })
			}
			if trackers, ok := tier.(*FallbackTrackersAnnouncer); ok {
				sort.Slice(trackers.trackers, func(i, j int) bool { return trackers.trackers[i].String() < trackers.trackers[j].String() })
			}
		}
	}
}
*/
