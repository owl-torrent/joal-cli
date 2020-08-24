package announce

//go:generate mockgen -destination=./announcer_mock.go -self_package=github.com/anthonyraymond/joal-cli/pkg/emulatedclient/announce -package=announce github.com/anthonyraymond/joal-cli/pkg/emulatedclient/announce IHttpAnnouncer,IUdpAnnouncer

import (
	"context"
	"fmt"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anthonyraymond/joal-cli/pkg/logs"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net"
	"net/url"
	"strings"
)

// tracker.AnnounceRequest uses a uint32 for the IPAddress, create our own struct that use proper net.IP type. UDP will need to convert this one to a tracker.AnnounceRequest
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

func (a *Announcer) Announce(u url.URL, announceRequest AnnounceRequest, ctx context.Context) (tracker.AnnounceResponse, error) {
	log := logs.GetLogger()
	var currentAnnouncer interface {
		Announce(url url.URL, announceRequest AnnounceRequest, ctx context.Context) (tracker.AnnounceResponse, error)
	}
	if strings.HasPrefix(u.Scheme, "http") {
		currentAnnouncer = a.Http
	} else if strings.HasPrefix(u.Scheme, "udp") {
		currentAnnouncer = a.Udp
	}

	if currentAnnouncer == nil { // some client file may not contains definitions for http or udp or the scheme might be a weird one
		return tracker.AnnounceResponse{}, errors.New(fmt.Sprintf("url='%s' => Scheme '%s' is not supported by the current client", u.String(), u.Scheme))
	}
	log.Info("announcing to tracker",
		zap.String("event", announceRequest.Event.String()),
		zap.ByteString("infohash", announceRequest.InfoHash[:]),
		zap.Int64("uploaded", announceRequest.Uploaded),
		zap.String("tracker", u.Host),
	)

	ret, err := currentAnnouncer.Announce(u, announceRequest, ctx)
	if err != nil {
		return tracker.AnnounceResponse{}, errors.Wrap(err, "failed to announce")
	}

	return ret, nil
}
