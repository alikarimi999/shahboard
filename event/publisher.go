package event

// Publisher is responsible for publishing events.
type Publisher interface {
	Publish(data ...Event) error
	Close() error
}
