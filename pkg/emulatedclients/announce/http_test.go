package announce

import (
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/casing"
	"github.com/anthonyraymond/joal-cli/pkg/emulatedclients/urlencoder"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"net"
	"net/http"
	"net/http/httptest"
	url2 "net/url"
	"testing"
	"text/template"
)

func TestHttpAnnouncer_ShouldUnmarshal(t *testing.T) {
	yamlString := `---
urlEncoder:
  encodedHexCase: lower
query: info_hash={{urlEncode (byteArray20ToString .InfoHash)}}&peer_id={{byteArray20ToString .PeerId}}&port={{.Port}}&uploaded={{.Uploaded}}&downloaded={{.Downloaded}}&left={{.Left}}&corrupt=0&key={{withLeadingZeroes (uint32ToHexString .Key) 8}}{{if ne .Event.String "empty"}}&event={{.Event.String}}{{end}}&numwant={{.NumWant}}&compact=1&no_peer_id=1&supportcrypto=1&redundant=0
requestHeaders:
  - name: User-Agent
    value: qBittorrent v3.3.1
  - name: Accept-Encoding
    value: gzip
`
	announcer := &HttpAnnouncer{}
	err := yaml.Unmarshal([]byte(yamlString), announcer)
	if err != nil {
		t.Fatalf("Failed to unmarshall: %+v", err)
	}
	assert.Equal(t, casing.Lower, announcer.UrlEncoder.EncodedHexCase)
	assert.Equal(t, `info_hash={{urlEncode (byteArray20ToString .InfoHash)}}&peer_id={{byteArray20ToString .PeerId}}&port={{.Port}}&uploaded={{.Uploaded}}&downloaded={{.Downloaded}}&left={{.Left}}&corrupt=0&key={{withLeadingZeroes (uint32ToHexString .Key) 8}}{{if ne .Event.String "empty"}}&event={{.Event.String}}{{end}}&numwant={{.NumWant}}&compact=1&no_peer_id=1&supportcrypto=1&redundant=0`, announcer.Query)
	assert.Equal(t, []HttpRequestHeader{{Name: "User-Agent", Value: "qBittorrent v3.3.1"}, {Name: "Accept-Encoding", Value: "gzip"}}, announcer.RequestHeaders)
}

func TestHttpAnnouncer_ShouldValidate(t *testing.T) {
	type args struct {
		Announcer HttpAnnouncer
	}
	tests := []struct {
		name         string
		args         args
		wantErr      bool
		failingField string
	}{
		{name: "shouldFailWithoutQuery", args: args{Announcer: HttpAnnouncer{Query: ""}}, wantErr: true, failingField: "Query"},
		{name: "shouldSucceedWithQuery", args: args{Announcer: HttpAnnouncer{Query: "dd"}}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.New().Struct(tt.args.Announcer)
			if tt.wantErr == true && err == nil {
				t.Fatal("validation failed, wantErr=true but err is nil")
			}
			if tt.wantErr == false && err != nil {
				t.Fatalf("validation failed, wantErr=false but err is : %v", err)
			}
			if tt.wantErr {
				validationErrors := err.(validator.ValidationErrors)
				fieldFound := false
				for _, e := range validationErrors {
					if e.Field() == tt.failingField {
						fieldFound = true
					}
				}
				if !fieldFound {
					t.Errorf("validation failed, field=%s not found in error list : %v", tt.failingField, validationErrors)
				}
			}
		})
	}
}

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


func TestHttpAnnouncer_AnnounceShouldAnnounce(t *testing.T) {
	expectedResponse := tracker.AnnounceResponse{
		Interval: 150,
		Leechers: 10,
		Seeders:  20,
		Peers:    tracker.Peers{tracker.Peer{IP: net.IPv4(10, 10, 10, 10), Port: 2501, ID: []byte{1}}},
	}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bytes, err := bencode.Marshal(expectedResponse)
		if err != nil {
			t.Errorf("failed to encode http announce response")
		}
		_, _ = w.Write(bytes)
	}))
	defer s.Close()

	announcer := HttpAnnouncer{
		UrlEncoder:     urlencoder.UrlEncoder{},
		Query:          "info_hash={{urlEncode (byteArray20ToString .InfoHash)}}",
		RequestHeaders: []HttpRequestHeader{},
	}
	err := announcer.AfterPropertiesSet()
	if err != nil {
		t.Fatal(err)
	}

	announceRequest := AnnounceRequest{
		InfoHash:   [20]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		PeerId:     [20]byte{},
		Downloaded: 0,
		Left:       0,
		Uploaded:   0,
		Event:      0,
		IPAddress:  nil,
		Key:        0,
		NumWant:    0,
		Port:       0,
	}
	url, _ := url2.Parse(s.URL)
	response, err := announcer.Announce(*url, announceRequest)
	if err != nil {
		t.Fatalf("httpAnnouncer.Announce() has failed: %v", err)
	}

	assert.Equal(t, response, response)
}

func TestHttpAnnouncer_Announce_ShouldNotPrefixWithAmpersandIfQueryHasNoValues(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedUrlSuffix := "/?info_hash=a"
		assert.Equal(t, expectedUrlSuffix, r.URL.String())

		_, _ = w.Write([]byte{})
	}))
	defer s.Close()

	announcer := HttpAnnouncer{
		UrlEncoder:     urlencoder.UrlEncoder{},
		Query:          "info_hash=a",
		RequestHeaders: []HttpRequestHeader{},
	}
	err := announcer.AfterPropertiesSet()
	if err != nil {
		t.Fatal(err)
	}

	url, _ := url2.Parse(s.URL)

	_, _ = announcer.Announce(*url, AnnounceRequest{})
}

func TestHttpAnnouncer_Announce_ShouldPrefixWithAmpersandQueryHasValues(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedUrlSuffix := "/?k=v&info_hash=a"
		assert.Equal(t, expectedUrlSuffix, r.URL.String())

		_, _ = w.Write([]byte{})
	}))
	defer s.Close()

	announcer := HttpAnnouncer{
		UrlEncoder:     urlencoder.UrlEncoder{},
		Query:          "info_hash=a",
		RequestHeaders: []HttpRequestHeader{},
	}
	err := announcer.AfterPropertiesSet()
	if err != nil {
		t.Fatal(err)
	}

	url, _ := url2.Parse(s.URL + "?k=v")

	_, _ = announcer.Announce(*url, AnnounceRequest{})
}

func TestHttpAnnouncer_Announce_ShouldSendHttpHeaders(t *testing.T) {
	headers := []HttpRequestHeader{{Name: "User-Agent", Value: "qBittorrent v3.3.1"}, {Name: "Accept-Encoding", Value: "gzip"}, {Name: "Last", Value: "happy"}}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Len(t, r.Header, len(headers))
		for _, expectedHeader := range headers {
			assert.Equal(t, expectedHeader.Value, r.Header.Get(expectedHeader.Name))
		}

		_, _ = w.Write([]byte{})
	}))
	defer s.Close()

	announcer := HttpAnnouncer{
		UrlEncoder:     urlencoder.UrlEncoder{},
		Query:          "info_hash=a",
		RequestHeaders: headers,
	}
	err := announcer.AfterPropertiesSet()
	if err != nil {
		t.Fatal(err)
	}

	url, _ := url2.Parse(s.URL + "?k=v")

	_, _ = announcer.Announce(*url, AnnounceRequest{})
}
