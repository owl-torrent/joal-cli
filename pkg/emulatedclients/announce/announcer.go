package announce

import (
	"fmt"
	"github.com/anacrolix/torrent/tracker"
	"github.com/pkg/errors"
	"net"
	"net/url"
	"strings"
)

type AnnounceRequest struct {
	InfoHash   [20]byte
	PeerId     [20]byte
	Downloaded int64
	Left       int64 // If less than 0, math.MaxInt64 will be used for HTTP trackers instead.
	Uploaded   int64
	// Apparently this is optional. None can be used for announces done at
	// regular intervals.
	Event     tracker.AnnounceEvent
	IPAddress net.IP
	Key       uint32
	NumWant   int32 // How many peer addresses are desired. -1 for default.
	Port      uint16
} // 82 bytes

type Announcer struct {
	Http IHttpAnnouncer `yaml:"http"`
	Udp  IUdpAnnouncer  `yaml:"udp"`
}

type AnnounceUrlList = []url.URL

func rotateLeft(l *AnnounceUrlList, offset int) {
	if offset < 0 || len(*l) == 0 {
		return
	}

	r := offset % len(*l)
	*l = append((*l)[r:], (*l)[:r]...)
}

func (a *Announcer) AfterPropertiesSet() error {
	if a.Http != nil {
		if err := a.Http.AfterPropertiesSet(); err != nil {
			return err
		}
	}
	if a.Udp != nil {
		if err := a.Udp.AfterPropertiesSet(); err != nil {
			return err
		}
	}
	return nil
}

// Announce to the announceURLs in order until one answer properly.
// The announceURLs array is modified in this method, a non answering tracker will be demoted to last position in the list.
// If none of the trackers respond the methods returns an error.
func (a *Announcer) Announce(announceURLs *[]url.URL, announceRequest AnnounceRequest) (ret tracker.AnnounceResponse, err error) {
	var tries = 0
	defer func(offset *int) { rotateLeft(announceURLs, *offset) }(&tries)

	var announceErrors = make([]error, 0)
	for i, announceUrl := range *announceURLs {
		tries = i
		var currentAnnouncer interface {
			Announce(url url.URL, announceRequest AnnounceRequest) (tracker.AnnounceResponse, error)
		}
		if strings.HasPrefix(announceUrl.Scheme, "http") {
			currentAnnouncer = a.Http
		} else if strings.HasPrefix(announceUrl.Scheme, "udp") {
			currentAnnouncer = a.Udp
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
