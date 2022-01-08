package announcer

import (
	"context"
	"fmt"
	"github.com/anacrolix/torrent/tracker"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// tracker.AnnounceRequest uses a uint32 for the IPAddress, create our own struct that use proper net.IP type. UDP will need to convert this one to a tracker.AnnounceRequest
type AnnounceRequest struct {
	InfoHash   [20]byte
	PeerId     [20]byte
	Downloaded int64
	Left       int64 // If less than 0, math.MaxInt64 will be used for HTTP trackers instead.
	Uploaded   int64
	Corrupt    int64
	// Apparently this is optional. None can be used for announces done at
	// regular intervals.
	Event     tracker.AnnounceEvent
	IPAddress net.IP
	Key       uint32
	NumWant   int32 // How many peer addresses are desired. -1 for default.
	Private   bool
	Port      uint16
} // 82 bytes

type AnnounceResponse struct {
	Interval time.Duration // Minimum seconds the local peer should wait before next announce.
	Leechers int32
	Seeders  int32
	Peers    []tracker.Peer
}

type iAnnouncer interface {
	Announce(u url.URL, announceRequest AnnounceRequest, ctx context.Context) (AnnounceResponse, error)
}

type Announcer struct {
	Http IHttpAnnouncer `yaml:"http" validate:"required_without_all=Udp"`
	Udp  IUdpAnnouncer  `yaml:"udp" validate:"required_without_all=Http"`
}

func (a *Announcer) UnmarshalYAML(value *yaml.Node) error {
	announcer := &struct {
		Http *HttpAnnouncer `yaml:"http"`
		// TODO: Udp *UdpAnnouncer `yaml:"udp"`
	}{}
	if a.Http != nil {
		announcer.Http = a.Http.(*HttpAnnouncer)
	}
	// TODO: add UDP
	/*if a.Udp != nil {
		announcer.Udp = a.Udp.(*UdpAnnouncer)
	}*/
	err := value.Decode(announcer)
	if err != nil {
		return err
	}

	(*a).Http = announcer.Http
	//TODO: (*a).Udp = announcer.udp

	return nil
}

func (a *Announcer) AfterPropertiesSet(proxyFunc func(*http.Request) (*url.URL, error)) error {
	if a.Http != nil {
		if err := a.Http.AfterPropertiesSet(proxyFunc); err != nil {
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

func (a *Announcer) Announce(u url.URL, announceRequest AnnounceRequest, ctx context.Context) (AnnounceResponse, error) {
	var currentAnnouncer interface {
		Announce(url url.URL, announceRequest AnnounceRequest, ctx context.Context) (AnnounceResponse, error)
	}
	if strings.HasPrefix(u.Scheme, "http") {
		currentAnnouncer = a.Http
	} else if strings.HasPrefix(u.Scheme, "udp") {
		currentAnnouncer = a.Udp
	}

	if currentAnnouncer == nil { // some client file may not contains definitions for http or udp or the scheme might be a weird one
		return AnnounceResponse{}, fmt.Errorf("url='%s' => Scheme '%s' is not supported by the current client", u.String(), u.Scheme)
	}

	ret, err := currentAnnouncer.Announce(u, announceRequest, ctx)
	if err != nil {
		return AnnounceResponse{}, errors.Wrap(err, "failed to announce")
	}

	return ret, nil
}
