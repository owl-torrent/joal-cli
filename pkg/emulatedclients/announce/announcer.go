package announce

import (
	"fmt"
	"github.com/anacrolix/torrent/tracker"
	"github.com/pkg/errors"
	"net/url"
	"strings"
)

type Announcer struct {
	http IHttpAnnouncer `yaml:"http"`
	udp  IUdpAnnouncer  `yaml:"udp"`
}

type AnnounceUrlList = []url.URL

func rotateLeft(l *AnnounceUrlList, offset int) {
	if offset < 0 || len(*l) == 0 {
		return
	}

	r := offset % len(*l)
	*l = append((*l)[r:], (*l)[:r]...)
}

// Announce to the announceURLs in order until one answer properly.
// The announceURLs array is modified in this method, a non answering tracker will be demoted to last position in the list.
// If none of the trackers respond the methods returns an error.
func (a *Announcer) Announce(announceURLs *[]url.URL, announceRequest tracker.AnnounceRequest) (ret tracker.AnnounceResponse, err error) {
	var tries = 0
	defer func(offset *int) { rotateLeft(announceURLs, *offset) }(&tries)

	var announceErrors = make([]error, 0)
	for i, announceUrl := range *announceURLs {
		tries = i
		var currentAnnouncer interface {
			Announce(url url.URL, announceRequest tracker.AnnounceRequest) (tracker.AnnounceResponse, error)
		}
		if strings.HasPrefix(announceUrl.Scheme, "http") {
			currentAnnouncer = a.http
		} else if strings.HasPrefix(announceUrl.Scheme, "udp") {
			currentAnnouncer = a.udp
		}

		if currentAnnouncer == nil { // some client file may not contains definitions for http or udp or the scheme might be a weird one
			announceErrors = append(announceErrors, errors.New(fmt.Sprintf("url='%s' => Scheme '%s' is not supported", announceUrl.String(), announceUrl.Scheme)))
			continue
		}

		ret, err = currentAnnouncer.Announce(announceUrl, announceRequest)
		if err == nil {
			return
		}

		announceErrors = append(announceErrors, err)
	}

	tries = len(*announceURLs)

	err = errors.New(fmt.Sprintf("failed to announce on every announce url: %+v", announceErrors))
	return
}

type MyType = *[]byte
