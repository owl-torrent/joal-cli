package events

type EventPublisher interface {
	Publish(interface{})
}
