package domain

type Disabled struct {
	Disabled bool
	Reason   string
}

var (
	announceProtocolNotSupported = Disabled{Disabled: true, Reason: "tracker.disabled.protocol-not-supported"}
	announceListNotSupported     = Disabled{Disabled: true, Reason: "tracker.disabled.announce-list-not-supported"}
)

type TrackerState struct {
	Disable          Disabled
	ConsecutiveFails int32
	StartSent        bool
	Updating         bool
}

func (s TrackerState) isDisabled() bool {
	return s.Disable.Disabled
}
