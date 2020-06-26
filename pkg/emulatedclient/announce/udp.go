package announce

import (
	"context"
	"github.com/anacrolix/torrent/tracker"
	"net/url"
)

type IUdpAnnouncer interface {
	Announce(url url.URL, announceRequest AnnounceRequest, ctx context.Context) (tracker.AnnounceResponse, error)
	AfterPropertiesSet() error
}
