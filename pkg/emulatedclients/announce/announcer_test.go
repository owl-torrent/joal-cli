package announce

import (
	"errors"
	"github.com/anacrolix/torrent/tracker"
	"github.com/stretchr/testify/assert"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func Test_rotateLeftX(t *testing.T) {
	type args struct {
		l      *AnnounceUrlList
		offset int
	}
	tests := []struct {
		name string
		args args
		want AnnounceUrlList
	}{
		{name: "shouldNotFailOnEmptyList", args: args{l: &AnnounceUrlList{}, offset: 3}, want: AnnounceUrlList{}},
		{name: "shouldNotFailOnSingleEntryList", args: args{l: &AnnounceUrlList{url.URL{Host: "a"}}, offset: 2}, want: AnnounceUrlList{url.URL{Host: "a"}}},
		{name: "shouldRotateZero", args: args{l: &AnnounceUrlList{url.URL{Host: "a"}, url.URL{Host: "b"}, url.URL{Host: "c"}}, offset: 0}, want: AnnounceUrlList{url.URL{Host: "a"}, url.URL{Host: "b"}, url.URL{Host: "c"}}},
		{name: "shouldRotateOne", args: args{l: &AnnounceUrlList{url.URL{Host: "a"}, url.URL{Host: "b"}, url.URL{Host: "c"}}, offset: 1}, want: AnnounceUrlList{url.URL{Host: "b"}, url.URL{Host: "c"}, url.URL{Host: "a"}}},
		{name: "shouldRotateExactArraySize", args: args{l: &AnnounceUrlList{url.URL{Host: "a"}, url.URL{Host: "b"}, url.URL{Host: "c"}}, offset: 3}, want: AnnounceUrlList{url.URL{Host: "a"}, url.URL{Host: "b"}, url.URL{Host: "c"}}},
		{name: "shouldRotateBiggerThanArraySize", args: args{l: &AnnounceUrlList{url.URL{Host: "a"}, url.URL{Host: "b"}, url.URL{Host: "c"}}, offset: 5}, want: AnnounceUrlList{url.URL{Host: "c"}, url.URL{Host: "a"}, url.URL{Host: "b"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rotateLeft(tt.args.l, tt.args.offset)
			if !reflect.DeepEqual(*tt.args.l, tt.want) {
				t.Errorf("rotateLeftX() = %v, want %v", *tt.args.l, tt.want)
			}
		})
	}
}

func TestAnnouncer_AnnounceShouldCallAnnouncerCorrespondingToScheme(t *testing.T) {
	announcer := Announcer{
		Http: &DumbHttpAnnouncer{},
		Udp:  &DumbUdpAnnouncer{},
	}

	_, _ = announcer.Announce(&[]url.URL{{Scheme: "http"}}, AnnounceRequest{})
	assert.Equal(t, 1, announcer.Http.(*DumbHttpAnnouncer).counter)

	_, _ = announcer.Announce(&[]url.URL{{Scheme: "https"}}, AnnounceRequest{})
	assert.Equal(t, 2, announcer.Http.(*DumbHttpAnnouncer).counter)

	_, _ = announcer.Announce(&[]url.URL{{Scheme: "udp"}}, AnnounceRequest{})
	assert.Equal(t, 1, announcer.Udp.(*DumbUdpAnnouncer).counter)

	_, _ = announcer.Announce(&[]url.URL{{Scheme: "udp4"}}, AnnounceRequest{})
	assert.Equal(t, 2, announcer.Udp.(*DumbUdpAnnouncer).counter)

	_, _ = announcer.Announce(&[]url.URL{{Scheme: "udp6"}}, AnnounceRequest{})
	assert.Equal(t, 3, announcer.Udp.(*DumbUdpAnnouncer).counter)
}

func TestAnnouncer_Announce_ShouldNotDemoteIfSucceed(t *testing.T) {
	announcer := Announcer{
		Http: &DumbHttpAnnouncer{},
		Udp:  &DumbUdpAnnouncer{},
	}

	urls := []url.URL{{Scheme: "http"}, {Scheme: "udp"}}
	_, _ = announcer.Announce(&urls, AnnounceRequest{})
	assert.Equal(t, 1, announcer.Http.(*DumbHttpAnnouncer).counter)
	assert.Equal(t, "http", urls[0].Scheme)
	assert.Equal(t, "udp", urls[1].Scheme)
	assert.Equal(t, 0, announcer.Udp.(*DumbUdpAnnouncer).counter)
}

func TestAnnouncer_Announce_ShouldDemoteFailingUrlsOnFail(t *testing.T) {
	announcer := Announcer{
		Http: &DumbHttpAnnouncer{},
		Udp:  &DumbUdpAnnouncer{},
	}

	urls := []url.URL{{Scheme: "http", Path: "fail"}, {Scheme: "http2", Path: "fail"}, {Scheme: "udp"}}
	_, _ = announcer.Announce(&urls, AnnounceRequest{})
	assert.Equal(t, 2, announcer.Http.(*DumbHttpAnnouncer).counter)
	assert.Equal(t, "udp", urls[0].Scheme)
	assert.Equal(t, "http", urls[1].Scheme)
	assert.Equal(t, "http2", urls[2].Scheme)
	assert.Equal(t, 1, announcer.Udp.(*DumbUdpAnnouncer).counter)
}

func TestAnnouncer_Announce_ShouldDemoteFailingUrlsOnFailAndReturnErrorIfNoneWorks(t *testing.T) {
	announcer := Announcer{
		Http: &DumbHttpAnnouncer{},
		Udp:  &DumbUdpAnnouncer{},
	}

	urls := []url.URL{{Scheme: "http", Path: "fail"}, {Scheme: "udp", Path: "fail"}}
	_, err := announcer.Announce(&urls, AnnounceRequest{})
	assert.NotNil(t, err)
	assert.Equal(t, 1, announcer.Http.(*DumbHttpAnnouncer).counter)
	assert.Equal(t, "http", urls[0].Scheme)
	assert.Equal(t, "udp", urls[1].Scheme)
	assert.Equal(t, 1, announcer.Udp.(*DumbUdpAnnouncer).counter)
}

type DumbHttpAnnouncer struct {
	counter int
}

func (a *DumbHttpAnnouncer) AfterPropertiesSet() error { return nil }
func (a *DumbHttpAnnouncer) Announce(url url.URL, announceRequest AnnounceRequest) (tracker.AnnounceResponse, error) {
	a.counter++
	if strings.Contains(url.String(), "fail") {
		return tracker.AnnounceResponse{}, errors.New("asked to fail because url contains 'fail'")
	}
	return tracker.AnnounceResponse{}, nil
}

type DumbUdpAnnouncer struct {
	counter int
}

func (a *DumbUdpAnnouncer) AfterPropertiesSet() error { return nil }
func (a *DumbUdpAnnouncer) Announce(url url.URL, announceRequest AnnounceRequest) (tracker.AnnounceResponse, error) {
	a.counter++
	if strings.Contains(url.String(), "fail") {
		return tracker.AnnounceResponse{}, errors.New("asked to fail because url contains 'fail'")
	}
	return tracker.AnnounceResponse{}, nil
}
