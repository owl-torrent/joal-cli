package announce

import (
	"github.com/anacrolix/torrent/tracker"
	"net/url"
)

type IUdpAnnouncer interface {
	Announce(url url.URL, announceRequest AnnounceRequest) (tracker.AnnounceResponse, error)
	AfterPropertiesSet() error
}
