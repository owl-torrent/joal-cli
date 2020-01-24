package announce

import (
	"github.com/anacrolix/torrent/metainfo"
	"reflect"
	"testing"
)

func Test_promoteTier(t *testing.T) {
	type args struct {
		al   metainfo.AnnounceList
		tier int
	}
	tests := []struct {
		name string
		args args
		want metainfo.AnnounceList
	}{
		{name: "shouldPromote", args: args{al: [][]string{{"t1"}}, tier: 0}, want: [][]string{{"t1"}}},
		{name: "shouldPromote", args: args{al: [][]string{{"t1"}, {"t2"}}, tier: 1}, want: [][]string{{"t2"}, {"t1"}}},
		{name: "shouldPromote", args: args{al: [][]string{{"t1"}, {"t2"}, {"t3"}}, tier: 0}, want: [][]string{{"t1"}, {"t2"}, {"t3"}}},
		{name: "shouldPromote", args: args{al: [][]string{{"t1"}, {"t2"}, {"t3"}}, tier: 1}, want: [][]string{{"t2"}, {"t1"}, {"t3"}}},
		{name: "shouldPromote", args: args{al: [][]string{{"t1"}, {"t2"}, {"t3"}}, tier: 2}, want: [][]string{{"t3"}, {"t1"}, {"t2"}}},
		{name: "shouldPromote", args: args{al: [][]string{{"t1"}, {"t2"}, {"t3"}, {"t4"}}, tier: 0}, want: [][]string{{"t1"}, {"t2"}, {"t3"}, {"t4"}}},
		{name: "shouldPromote", args: args{al: [][]string{{"t1"}, {"t2"}, {"t3"}, {"t4"}}, tier: 1}, want: [][]string{{"t2"}, {"t1"}, {"t3"}, {"t4"}}},
		{name: "shouldPromote", args: args{al: [][]string{{"t1"}, {"t2"}, {"t3"}, {"t4"}}, tier: 2}, want: [][]string{{"t3"}, {"t1"}, {"t2"}, {"t4"}}},
		{name: "shouldPromote", args: args{al: [][]string{{"t1"}, {"t2"}, {"t3"}, {"t4"}}, tier: 3}, want: [][]string{{"t4"}, {"t1"}, {"t2"}, {"t3"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			promoteTier(&tt.args.al, tt.args.tier)
			if !reflect.DeepEqual(tt.args.al, tt.want) {
				t.Errorf("promoteTier() = %v, want %v", tt.args.al, tt.want)
			}
		})
	}
}

func Test_promoteUrlInTier(t *testing.T) {
	type args struct {
		al       metainfo.AnnounceList
		tier     int
		urlIndex int
	}
	tests := []struct {
		name string
		args args
		want metainfo.AnnounceList
	}{
		{name: "shouldPromote", args: args{al: [][]string{{"t1.1"}}, tier: 0, urlIndex: 0}, want: [][]string{{"t1.1"}}},
		{name: "shouldPromote", args: args{al: [][]string{{"t1.1", "t1.2"}}, tier: 0, urlIndex: 0}, want: [][]string{{"t1.1", "t1.2"}}},
		{name: "shouldPromote", args: args{al: [][]string{{"t1.1", "t1.2"}}, tier: 0, urlIndex: 1}, want: [][]string{{"t1.2", "t1.1"}}},
		{name: "shouldPromote", args: args{al: [][]string{{"t1.1", "t1.2", "t1.3"}}, tier: 0, urlIndex: 0}, want: [][]string{{"t1.1", "t1.2", "t1.3"}}},
		{name: "shouldPromote", args: args{al: [][]string{{"t1.1", "t1.2", "t1.3"}}, tier: 0, urlIndex: 1}, want: [][]string{{"t1.2", "t1.1", "t1.3"}}},
		{name: "shouldPromote", args: args{al: [][]string{{"t1.1", "t1.2", "t1.3"}}, tier: 0, urlIndex: 2}, want: [][]string{{"t1.3", "t1.1", "t1.2"}}},
		{name: "shouldPromote", args: args{al: [][]string{{"t1.1", "t1.2", "t1.3"}, {"t2.1"}}, tier: 0, urlIndex: 0}, want: [][]string{{"t1.1", "t1.2", "t1.3"}, {"t2.1"}}},
		{name: "shouldPromote", args: args{al: [][]string{{"t1.1", "t1.2", "t1.3"}, {"t2.1"}}, tier: 0, urlIndex: 1}, want: [][]string{{"t1.2", "t1.1", "t1.3"}, {"t2.1"}}},
		{name: "shouldPromote", args: args{al: [][]string{{"t1.1", "t1.2", "t1.3"}, {"t2.1"}}, tier: 0, urlIndex: 2}, want: [][]string{{"t1.3", "t1.1", "t1.2"}, {"t2.1"}}},
		{name: "shouldPromote", args: args{al: [][]string{{"t1.1", "t1.2", "t1.3"}, {"t2.1"}}, tier: 1, urlIndex: 0}, want: [][]string{{"t1.1", "t1.2", "t1.3"}, {"t2.1"}}},
		{name: "shouldPromote", args: args{al: [][]string{{"t1.1", "t1.2", "t1.3"}, {"t2.1", "t2.2"}}, tier: 0, urlIndex: 0}, want: [][]string{{"t1.1", "t1.2", "t1.3"}, {"t2.1", "t2.2"}}},
		{name: "shouldPromote", args: args{al: [][]string{{"t1.1", "t1.2", "t1.3"}, {"t2.1", "t2.2"}}, tier: 0, urlIndex: 1}, want: [][]string{{"t1.2", "t1.1", "t1.3"}, {"t2.1", "t2.2"}}},
		{name: "shouldPromote", args: args{al: [][]string{{"t1.1", "t1.2", "t1.3"}, {"t2.1", "t2.2"}}, tier: 0, urlIndex: 2}, want: [][]string{{"t1.3", "t1.1", "t1.2"}, {"t2.1", "t2.2"}}},
		{name: "shouldPromote", args: args{al: [][]string{{"t1.1", "t1.2", "t1.3"}, {"t2.1", "t2.2"}}, tier: 1, urlIndex: 0}, want: [][]string{{"t1.1", "t1.2", "t1.3"}, {"t2.1", "t2.2"}}},
		{name: "shouldPromote", args: args{al: [][]string{{"t1.1", "t1.2", "t1.3"}, {"t2.1", "t2.2"}}, tier: 1, urlIndex: 1}, want: [][]string{{"t1.1", "t1.2", "t1.3"}, {"t2.2", "t2.1"}}},
		{name: "shouldPromote", args: args{al: [][]string{{"t1.1", "t1.2", "t1.3"}, {"t2.1", "t2.2", "t2.3"}}, tier: 1, urlIndex: 0}, want: [][]string{{"t1.1", "t1.2", "t1.3"}, {"t2.1", "t2.2", "t2.3"}}},
		{name: "shouldPromote", args: args{al: [][]string{{"t1.1", "t1.2", "t1.3"}, {"t2.1", "t2.2", "t2.3"}}, tier: 1, urlIndex: 1}, want: [][]string{{"t1.1", "t1.2", "t1.3"}, {"t2.2", "t2.1", "t2.3"}}},
		{name: "shouldPromote", args: args{al: [][]string{{"t1.1", "t1.2", "t1.3"}, {"t2.1", "t2.2", "t2.3"}}, tier: 1, urlIndex: 2}, want: [][]string{{"t1.1", "t1.2", "t1.3"}, {"t2.3", "t2.1", "t2.2"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			promoteUrlInTier(&tt.args.al, tt.args.tier, tt.args.urlIndex)
			if !reflect.DeepEqual(tt.args.al, tt.want) {
				t.Errorf("promoteUrlInTier() = %v, want %v", tt.args.al, tt.want)
			}
		})
	}
}
