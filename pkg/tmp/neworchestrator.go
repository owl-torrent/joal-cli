package tmp

import (
	"github.com/anacrolix/torrent/tracker"
	"github.com/google/uuid"
)

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
