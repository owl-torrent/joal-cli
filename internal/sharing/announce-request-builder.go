package sharing

import "net/url"

type announceRequestBuilder struct {
	url   *url.URL
	event AnnounceEvent
}

func (b *announceRequestBuilder) withUrl(u *url.URL) {
	b.url = u
}

func (b *announceRequestBuilder) withEvent(event AnnounceEvent) {
	b.event = event
}
