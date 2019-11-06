package announce

import (
	"github.com/anacrolix/torrent/tracker"
	"net/url"
)

type IHttpAnnouncer interface {
	Announce(url url.URL, announceRequest tracker.AnnounceRequest) (tracker.AnnounceResponse, error)
}
