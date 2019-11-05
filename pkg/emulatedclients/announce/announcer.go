package announce

import (
	"fmt"
	"github.com/anacrolix/torrent/tracker"
	"github.com/pkg/errors"
	"net/url"
	"strings"
)

type Announcer struct {
	Numwant       int            `yaml:"numwant"`
	NumwantOnStop int            `yaml:"numwantOnStop"`
	http          IHttpAnnouncer `yaml:"http"`
	udp           IUdpAnnouncer  `yaml:"udp"`
}

type IAnnounceAble interface {
	Downloaded() int64
	Uploaded() int64
	Left() int64
}

type AnnounceUrlList = []url.URL

func rotateLeft(l *AnnounceUrlList) {
	listLen := len(*l)
	if listLen < 2 {
		return
	}
	*l = append((*l)[1:], (*l)[0])
}

func (a *Announcer) Announce(announceUrls *[]url.URL, iAnnounceAble IAnnounceAble) (*tracker.AnnounceResponse, error) {
	// iterate on a copy
	var iterableUrls = make(AnnounceUrlList, len(*announceUrls))
	copy(iterableUrls, *announceUrls)
	var announceErrors = make([]error, 0)
	for _, announceUrl := range iterableUrls {
		var currentAnnouncer interface {
			Announce(url url.URL, iAnnounceAble IAnnounceAble) (*tracker.AnnounceResponse, error)
		}
		if strings.HasPrefix(announceUrl.Scheme, "http") {
			currentAnnouncer = a.http
		} else if strings.HasPrefix(announceUrl.Scheme, "udp") {
			currentAnnouncer = a.udp
		} else {
			return nil, errors.New(fmt.Sprintf("Scheme '%s' is not supported", announceUrl.Scheme))
		}

		if currentAnnouncer == nil { // some client file may not contains definitions for http or udp
			announceErrors = append(announceErrors, errors.New(fmt.Sprintf("Client does not support '%s' protocol", announceUrl.Scheme)))
			rotateLeft(announceUrls)
			continue
		}

		res, err := currentAnnouncer.Announce(announceUrl, iAnnounceAble)
		if err == nil {
			return res, nil
		}

		rotateLeft(announceUrls)
		announceErrors = append(announceErrors, err)
	}

	return nil, errors.New(fmt.Sprintf("failed to announce on every announce url: %+v", announceErrors))
}

type MyType = *[]byte
