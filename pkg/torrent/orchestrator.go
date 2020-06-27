package torrent

//go:generate mockgen -destination=./orchestrator_mock.go -self_package=github.com/anthonyraymond/joal-cli/pkg/torrent -package=torrent github.com/anthonyraymond/joal-cli/pkg/torrent ITrackerAnnouncer,ITierAnnouncer

import (
	"context"
	"github.com/anacrolix/torrent/tracker"
	"github.com/google/uuid"
	"net/url"
)

type AnnouncingFunction = func(u url.URL, event tracker.AnnounceEvent, ctx context.Context) trackerAnnounceResult

type ITrackerAnnouncer interface {
	Uuid() uuid.UUID
	announceOnce(announce AnnouncingFunction, event tracker.AnnounceEvent) trackerAwareAnnounceResult
	startAnnounceLoop(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent)
	Responses() <-chan trackerAwareAnnounceResult
	stopAnnounceLoop()
}

type ITierAnnouncer interface {
	Uuid() uuid.UUID
	announceOnce(announce AnnouncingFunction, event tracker.AnnounceEvent) tierState
	startAnnounceLoop(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent)
	States() <-chan tierState
	stopAnnounceLoop()
}
