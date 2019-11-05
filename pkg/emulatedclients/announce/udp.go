package announce

import (
	"github.com/anacrolix/torrent/tracker"
	"net/url"
)

type IUdpAnnouncer interface {
	Announce(url url.URL, iAnnounceAble IAnnounceAble) (*tracker.AnnounceResponse, error)
}
