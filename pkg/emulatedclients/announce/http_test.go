package announce

import (
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/casing"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/urlencoder"
	"net"
	"testing"
	"text/template"
)

func Test_setupQuery(t *testing.T) {
	urlEncoder := urlencoder.UrlEncoder{EncodedHexCase: casing.Upper}

	type args struct {
		templateStr     string
		announceRequest AnnounceRequest
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "shouldReplaceInfoHash", args: args{templateStr: "infohash={{urlEncode (byteArray20ToString .InfoHash)}}", announceRequest: AnnounceRequest{InfoHash: [20]byte{0x61, 0x61, 0x00, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61}}}, want: "infohash=aa%00aaaaaaaaaaaaaaaaa"},
		{name: "shouldReplacePeerid", args: args{templateStr: "peerid={{byteArray20ToString .PeerId}}", announceRequest: AnnounceRequest{PeerId: [20]byte{0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61}}}, want: "peerid=aaaaaaaaaaaaaaaaaaaa"},
		{name: "shouldReplaceKey", args: args{templateStr: "key={{uint32ToHexString .Key}}", announceRequest: AnnounceRequest{Key: 10}}, want: "key=a"},
		{name: "shouldReplacePort", args: args{templateStr: "port={{.Port}}", announceRequest: AnnounceRequest{Port: 2561}}, want: "port=2561"},
		{name: "shouldReplaceUploaded", args: args{templateStr: "uploaded={{.Uploaded}}", announceRequest: AnnounceRequest{Uploaded: 2561}}, want: "uploaded=2561"},
		{name: "shouldReplaceDownloaded", args: args{templateStr: "downloaded={{.Downloaded}}", announceRequest: AnnounceRequest{Downloaded: 2561}}, want: "downloaded=2561"},
		{name: "shouldReplaceLeft", args: args{templateStr: "left={{.Left}}", announceRequest: AnnounceRequest{Left: 2561}}, want: "left=2561"},
		{name: "shouldReplaceNumWant", args: args{templateStr: "numwant={{.NumWant}}", announceRequest: AnnounceRequest{NumWant: 30}}, want: "numwant=30"},
		{name: "shouldReplaceEvent", args: args{templateStr: "{{ if ne .Event.String \"empty\"}}&event={{.Event.String}}{{end}}", announceRequest: AnnounceRequest{Event: tracker.Started}}, want: "&event=started"},
		{name: "shouldReplaceEvent", args: args{templateStr: "{{ if ne .Event.String \"empty\"}}&event={{.Event.String}}{{end}}", announceRequest: AnnounceRequest{Event: tracker.Stopped}}, want: "&event=stopped"},
		{name: "shouldReplaceEvent", args: args{templateStr: "{{ if ne .Event.String \"empty\"}}&event={{.Event.String}}{{end}}", announceRequest: AnnounceRequest{Event: tracker.Completed}}, want: "&event=completed"},
		{name: "shouldReplaceEvent", args: args{templateStr: "{{ if ne .Event.String \"empty\"}}&event={{.Event.String}}{{end}}", announceRequest: AnnounceRequest{Event: tracker.None}}, want: ""},
		{name: "shouldBuildCompleteUrl", args: args{
			templateStr: "info_hash={{urlEncode (byteArray20ToString .InfoHash)}}&peer_id={{byteArray20ToString .PeerId}}&port={{.Port}}&uploaded={{.Uploaded}}&downloaded={{.Downloaded}}&left={{.Left}}&corrupt=0&key={{withLeadingZeroes (uint32ToHexString .Key) 8}}{{if ne .Event.String \"empty\"}}&event={{.Event.String}}{{end}}&numwant={{.NumWant}}&compact=1&no_peer_id=1&supportcrypto=1&redundant=0",
			announceRequest: AnnounceRequest{
				InfoHash:   [20]byte{0x61, 0x61, 0x00, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61, 0x61},
				PeerId:     [20]byte{0x62, 0x62, 0x62, 0x62, 0x62, 0x62, 0x62, 0x62, 0x62, 0x62, 0x62, 0x62, 0x62, 0x62, 0x62, 0x62, 0x62, 0x62, 0x62, 0x62},
				Downloaded: 5200,
				Left:       58,
				Uploaded:   1234568,
				Event:      tracker.Started,
				IPAddress:  net.IPv4zero,
				Key:        uint32(13),
				NumWant:    30,
				Port:       14598,
			},
		}, want: "info_hash=aa%00aaaaaaaaaaaaaaaaa&peer_id=bbbbbbbbbbbbbbbbbbbb&port=14598&uploaded=1234568&downloaded=5200&left=58&corrupt=0&key=0000000d&event=started&numwant=30&compact=1&no_peer_id=1&supportcrypto=1&redundant=0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := template.Must(template.New("test-tmpl").Funcs(TemplateFunctions(&urlEncoder)).Parse(tt.args.templateStr))
			got, err := buildQueryString(tmpl, tt.args.announceRequest)
			if err != nil {
				t.Errorf("buildQueryString() failed: %+v", err)
			}
			if got != tt.want {
				t.Errorf("buildQueryString() = %v, want %v", got, tt.want)
			}
		})
	}
}
