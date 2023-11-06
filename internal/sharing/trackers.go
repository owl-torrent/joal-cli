package sharing

type Trackers interface {
}

type trackersList struct {
}

func newTrackers() Trackers {
	return &trackersList{}
}
