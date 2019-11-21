package announce

import (
	"errors"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/tracker"
	"github.com/stretchr/testify/assert"
	"net/url"
	"strings"
	"testing"
)

func TestAnnouncer_AnnounceShouldCallAnnouncerCorrespondingToScheme(t *testing.T) {
	announcer := Announcer{
		Http: &DumbHttpAnnouncer{},
		Udp:  &DumbUdpAnnouncer{},
	}

	_, _ = announcer.Announce(&metainfo.AnnounceList{{"http://localhost.fr"}}, AnnounceRequest{})
	assert.Equal(t, 1, announcer.Http.(*DumbHttpAnnouncer).counter)

	_, _ = announcer.Announce(&metainfo.AnnounceList{{"https://localhost.fr"}}, AnnounceRequest{})
	assert.Equal(t, 2, announcer.Http.(*DumbHttpAnnouncer).counter)

	_, _ = announcer.Announce(&metainfo.AnnounceList{{"udp://localhost.fr"}}, AnnounceRequest{})
	assert.Equal(t, 1, announcer.Udp.(*DumbUdpAnnouncer).counter)

	_, _ = announcer.Announce(&metainfo.AnnounceList{{"udp4://localhost.fr"}}, AnnounceRequest{})
	assert.Equal(t, 2, announcer.Udp.(*DumbUdpAnnouncer).counter)

	_, _ = announcer.Announce(&metainfo.AnnounceList{{"udp6://localhost.fr"}}, AnnounceRequest{})
	assert.Equal(t, 3, announcer.Udp.(*DumbUdpAnnouncer).counter)
}

func TestAnnouncer_Announce_ShouldNotDemoteIfSucceed(t *testing.T) {
	announcer := Announcer{
		Http: &DumbHttpAnnouncer{},
		Udp:  &DumbUdpAnnouncer{},
	}

	urls := metainfo.AnnounceList{{"http://localhost.fr", "udp://localhost.fr"}}
	expected := metainfo.AnnounceList{{"http://localhost.fr", "udp://localhost.fr"}}
	_, _ = announcer.Announce(&urls, AnnounceRequest{})
	assert.Equal(t, 1, announcer.Http.(*DumbHttpAnnouncer).counter)
	assert.Equal(t, expected, urls)
	assert.Equal(t, 0, announcer.Udp.(*DumbUdpAnnouncer).counter)
}

func TestAnnouncer_Announce_ShouldPromoteTierAndUrlInTierIfSucceed(t *testing.T) {
	announcer := Announcer{
		Http: &DumbHttpAnnouncer{},
		Udp:  &DumbUdpAnnouncer{},
	}

	urls := metainfo.AnnounceList{
		{"http://localhost.fr/fail", "http://localhost.fr/x/fail", "http://localhost.fr/y/fail"},
		{"http://localhost.fr/t2/fail", "http://localhost.fr/t2/x/fail", "http://localhost.fr/t2/y"},
	}
	expected := metainfo.AnnounceList{
		{"http://localhost.fr/t2/y", "http://localhost.fr/t2/fail", "http://localhost.fr/t2/x/fail"},
		{"http://localhost.fr/fail", "http://localhost.fr/x/fail", "http://localhost.fr/y/fail"},
	}
	_, _ = announcer.Announce(&urls, AnnounceRequest{})
	assert.Equal(t, 6, announcer.Http.(*DumbHttpAnnouncer).counter)
	assert.Equal(t, expected, urls)
	assert.Equal(t, 0, announcer.Udp.(*DumbUdpAnnouncer).counter)
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
