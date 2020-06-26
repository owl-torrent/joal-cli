package announce

import (
	"context"
	"fmt"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/tracker"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
	Http IHttpAnnouncer `yaml:"http" validate:"required_without_all=Udp"`
	Udp  IUdpAnnouncer  `yaml:"udp" validate:"required_without_all=Http"`
}

func (a *Announcer) UnmarshalYAML(unmarshal func(interface{}) error) error {
	announcer := &struct {
		Http HttpAnnouncer `yaml:"http"`
		// TODO: Udp UdpAnnouncer `yaml:"udp"`
	}{}
	err := unmarshal(&announcer)
	if err != nil {
		return err
	}

	(*a).Http = &announcer.Http
	//TODO: (*a).Udp = &udp

	return nil
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
func (a *Announcer) Announce(announceUrlList *metainfo.AnnounceList, announceRequest AnnounceRequest, ctx context.Context) (tracker.AnnounceResponse, error) {
	var announceErrors = make([]error, 0)

	for iTier, tier := range *announceUrlList {
		for iUrl, uri := range tier {
			announceUrl, err := url.Parse(uri)
			if err != nil {
				announceErrors = append(announceErrors, errors.New(fmt.Sprintf("failed to parse announce url '%s': %v", uri, err)))
				continue
			}
			var currentAnnouncer interface {
				Announce(url url.URL, announceRequest AnnounceRequest, ctx context.Context) (tracker.AnnounceResponse, error)
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

			logrus.
				WithField("event", announceRequest.Event).
				WithField("infohash", announceRequest.InfoHash).
				WithField("uploaded", announceRequest.Uploaded).
				Info("announcing to tracker")

			ret, err := currentAnnouncer.Announce(*announceUrl, announceRequest, ctx)
			if err == nil {
				promoteUrlInTier(announceUrlList, iTier, iUrl)
				promoteTier(announceUrlList, iTier)
				return ret, nil
			}

			announceErrors = append(announceErrors, err)
		}
	}

	err := errors.New(fmt.Sprintf("failed to announce on every announce url: %+v", announceErrors))
	return tracker.AnnounceResponse{}, err
}
