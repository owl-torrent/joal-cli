package tmp

/*
type ITierAnnouncer interface {
	loop(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent)
	stopLoop()
}

type AllTrackersTierAnnouncer struct {
	trackers []trackerAnnouncer
	state chan tierState
	stopChan chan chan struct{}
}


func (t *AllTrackersTierAnnouncer) loop(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) {
	for _, tr := range t.trackers {
		go tr.startAnnounceLoop(announce, firstEvent)
	}

	for {
		select {
		case

		}
	}
}

func (t *AllTrackersTierAnnouncer) stopLoop() {
	done := make(chan struct{})
	t.stopChan <- done
	<- done
}

type FallbackTrackersTierAnnouncer struct {
	trackers []trackerAnnouncer
	stopChan chan chan struct{}
}


func (t *FallbackTrackersTierAnnouncer) loop(announce AnnouncingFunction, firstEvent tracker.AnnounceEvent) {

}

func (t *FallbackTrackersTierAnnouncer) stopLoop() {
	done := make(chan struct{})
	t.stopChan <- done
	<- done
}


type tierState = byte
const (
	ALIVE tierState = iota
	DEAD            = 0x01
)*/
