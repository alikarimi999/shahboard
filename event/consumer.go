package event

// Subscriber represents an interface for managing a subscription to a topic and interacting with a consumer.
type Subscriber interface {
	Subscribe(topic Topic) Subscription // Subscribes to a topic and returns a Subscription instance.
	Close() error                       // Closes the subscriber and releases resources.
}
