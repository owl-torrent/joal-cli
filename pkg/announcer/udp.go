package announcer

type IUdpAnnouncer interface {
	iAnnouncer
	AfterPropertiesSet() error
}
