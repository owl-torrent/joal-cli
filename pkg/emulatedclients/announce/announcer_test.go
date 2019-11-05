package announce

import (
	"errors"
	"github.com/anacrolix/torrent/tracker"
	"github.com/stretchr/testify/assert"
	"net/url"
	"strings"
	"testing"
)

func TestAnnounceUrlList_DemoteShouldNotFailOnEmptyList(t *testing.T) {
	list := AnnounceUrlList{}

	rotateLeft(&list)
	assert.Empty(t, list)
}
func TestAnnounceUrlList_DemoteShouldNotFailOnSingleEntryList(t *testing.T) {
	list := AnnounceUrlList{url.URL{Host: "a"}}

	rotateLeft(&list)
	assert.Len(t, list, 1)
}

func TestAnnounceUrlList_DemoteShouldDemote(t *testing.T) {
	list := AnnounceUrlList{url.URL{Host: "a"}, url.URL{Host: "b"}, url.URL{Host: "c"}}

	rotateLeft(&list)
	assert.Equal(t, url.URL{Host: "b"}, list[0])
	assert.Equal(t, url.URL{Host: "c"}, list[1])
	assert.Equal(t, url.URL{Host: "a"}, list[2])

	rotateLeft(&list)
	assert.Equal(t, url.URL{Host: "c"}, list[0])
	assert.Equal(t, url.URL{Host: "a"}, list[1])
	assert.Equal(t, url.URL{Host: "b"}, list[2])

	rotateLeft(&list)
	assert.Equal(t, url.URL{Host: "a"}, list[0])
	assert.Equal(t, url.URL{Host: "b"}, list[1])
	assert.Equal(t, url.URL{Host: "c"}, list[2])
}

func TestAnnouncer_AnnounceShouldCallAnnouncerCorrespondingToScheme(t *testing.T) {
	announcer := Announcer{
		Numwant:       0,
		NumwantOnStop: 0,
		http:          &DumbHttpAnnouncer{},
		udp:           &DumbUdpAnnouncer{},
	}

	announceAble := &DumbAnnounceAble{}

	_, _ = announcer.Announce(&[]url.URL{{Scheme: "http"}}, announceAble)
	assert.Equal(t, 1, announcer.http.(*DumbHttpAnnouncer).counter)

	_, _ = announcer.Announce(&[]url.URL{{Scheme: "https"}}, announceAble)
	assert.Equal(t, 2, announcer.http.(*DumbHttpAnnouncer).counter)

	_, _ = announcer.Announce(&[]url.URL{{Scheme: "udp"}}, announceAble)
	assert.Equal(t, 1, announcer.udp.(*DumbUdpAnnouncer).counter)

	_, _ = announcer.Announce(&[]url.URL{{Scheme: "udp4"}}, announceAble)
	assert.Equal(t, 2, announcer.udp.(*DumbUdpAnnouncer).counter)

	_, _ = announcer.Announce(&[]url.URL{{Scheme: "udp6"}}, announceAble)
	assert.Equal(t, 3, announcer.udp.(*DumbUdpAnnouncer).counter)
}

func TestAnnouncer_ShouldTryNextUrlWhenFailsAndPromote(t *testing.T) {
	announcer := Announcer{
		Numwant:       0,
		NumwantOnStop: 0,
		http:          &DumbHttpAnnouncer{},
		udp:           &DumbUdpAnnouncer{},
	}

	announceAble := &DumbAnnounceAble{}

	urls := []url.URL{{Scheme: "http", Path: "fail"}, {Scheme: "udp"}}
	_, _ = announcer.Announce(&urls, announceAble)
	assert.Equal(t, 1, announcer.http.(*DumbHttpAnnouncer).counter)
	assert.Equal(t, "udp", urls[0].Scheme)
	assert.Equal(t, "http", urls[1].Scheme)
	assert.Equal(t, 1, announcer.udp.(*DumbUdpAnnouncer).counter)
}

func TestAnnouncer_ShouldTryNextUrlWhenFailsAndReturnErrorIfNoneWorks(t *testing.T) {
	announcer := Announcer{
		Numwant:       0,
		NumwantOnStop: 0,
		http:          &DumbHttpAnnouncer{},
		udp:           &DumbUdpAnnouncer{},
	}

	announceAble := &DumbAnnounceAble{}

	urls := []url.URL{{Scheme: "http", Path: "fail"}, {Scheme: "udp", Path: "fail"}}
	_, err := announcer.Announce(&urls, announceAble)
	assert.NotNil(t, err)
	assert.Equal(t, 1, announcer.http.(*DumbHttpAnnouncer).counter)
	assert.Equal(t, "http", urls[0].Scheme)
	assert.Equal(t, "udp", urls[1].Scheme)
	assert.Equal(t, 1, announcer.udp.(*DumbUdpAnnouncer).counter)
}

type DumbAnnounceAble struct {
	announceURL *AnnounceUrlList // take a pointer to modify the slice (promote / demote url)
	downloaded  int64
	uploaded    int64
	left        int64
}

func (aa *DumbAnnounceAble) AnnounceURL() *AnnounceUrlList { return aa.announceURL }
func (aa *DumbAnnounceAble) Downloaded() int64             { return aa.downloaded }
func (aa *DumbAnnounceAble) Uploaded() int64               { return aa.uploaded }
func (aa *DumbAnnounceAble) Left() int64                   { return aa.left }

type DumbHttpAnnouncer struct {
	counter int
}

func (a *DumbHttpAnnouncer) Announce(url url.URL, iAnnounceAble IAnnounceAble) (*tracker.AnnounceResponse, error) {
	a.counter++
	if strings.Contains(url.String(), "fail") {
		return nil, errors.New("asked to fail because url contains 'fail'")
	}
	return nil, nil
}

type DumbUdpAnnouncer struct {
	counter int
}

func (a *DumbUdpAnnouncer) Announce(url url.URL, iAnnounceAble IAnnounceAble) (*tracker.AnnounceResponse, error) {
	a.counter++
	if strings.Contains(url.String(), "fail") {
		return nil, errors.New("asked to fail because url contains 'fail'")
	}
	return nil, nil
}
