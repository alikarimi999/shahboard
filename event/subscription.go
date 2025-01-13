package event

// Subscription represents an interface for managing a single subscription to a topic.
type Subscription interface {
	Topic() Topic        // Returns the topic associated with the subscription.
	Event() <-chan Event // Returns a channel to receive events.
	Err() <-chan error   // Returns a channel to receive errors.
	Unsubscribe()        // Unsubscribes from the topic and stops receiving events.
}
