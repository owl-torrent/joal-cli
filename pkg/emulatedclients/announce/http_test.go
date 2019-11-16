package announce

import (
	"github.com/anacrolix/torrent/tracker"
	"net/url"
	"testing"
)

func Test_setupQuery(t *testing.T) {
	type args struct {
		url            url.URL
		announceRequest tracker.AnnounceRequest
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name:"shouldReplaceInfoHash", args:args{url: url.URL{RawQuery:"infohash={infohash}"}, announceRequest: tracker.AnnounceRequest{InfoHash: [20]byte{0x61,0x61,0x61,0x61,0x61,0x61,0x61,0x61,0x61,0x61,0x61,0x61,0x61,0x61,0x61,0x61,0x61,0x61,0x61,0x61}}}, want:"infohash=aaaaaaaaaaaaaaaaaaaa"},
		{name:"shouldReplacePeerid", args:args{url: url.URL{RawQuery:"peerid={peerid}"}, announceRequest: tracker.AnnounceRequest{PeerId: [20]byte{0x61,0x61,0x61,0x61,0x61,0x61,0x61,0x61,0x61,0x61,0x61,0x61,0x61,0x61,0x61,0x61,0x61,0x61,0x61,0x61}}}, want:"peerid=aaaaaaaaaaaaaaaaaaaa"},
		{name:"shouldReplacePort", args:args{url: url.URL{RawQuery:"port={port}"}, announceRequest: tracker.AnnounceRequest{Port:2561}}, want:"port=2561"},
		{name:"shouldReplaceUploaded", args:args{url: url.URL{RawQuery:"uploaded={uploaded}"}, announceRequest: tracker.AnnounceRequest{Uploaded:2561}}, want:"uploaded=2561"},
		{name:"shouldReplaceDownloaded", args:args{url: url.URL{RawQuery:"downloaded={downloaded}"}, announceRequest: tracker.AnnounceRequest{Downloaded:2561}}, want:"downloaded=2561"},
		{name:"shouldReplaceLeft", args:args{url: url.URL{RawQuery:"left={left}"}, announceRequest: tracker.AnnounceRequest{Left:2561}}, want:"left=2561"},
		{name:"shouldReplaceKey", args:args{url: url.URL{RawQuery:"key={key}"}, announceRequest: tracker.AnnounceRequest{Key:12}}, want:"A"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupQuery(&tt.args.url, tt.args.announceRequest);
			got := tt.args.url.RawQuery
			if got != tt.want {
				t.Errorf("setupQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}