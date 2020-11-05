package announcer

import (
	"context"
	"net/url"
)

type IUdpAnnouncer interface {
	Announce(url url.URL, announceRequest AnnounceRequest, ctx context.Context) (AnnounceResponse, error)
	AfterPropertiesSet() error
}
