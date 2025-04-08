package bot

import (
	"fmt"
	"sync"
)

type Topic string
type Event struct {
	Topic Topic
	Data  interface{}
}

// PubSub handles message dispatching to workers
type PubSub struct {
	subscribers map[Topic]map[int]*Subscription
	mu          sync.RWMutex
}

// NewPubSub initializes a PubSub instance
func NewPubSub() *PubSub {
	return &PubSub{
		subscribers: make(map[Topic]map[int]*Subscription),
	}
}

func (u *Bot) Publish(e Event) {
	u.ps.publish(e)
}

func (u *Bot) Subscribe(t Topic) *Subscription {
	return u.ps.subscribe(t)
}

func (ps *PubSub) subscribe(t Topic) *Subscription {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	id := len(ps.subscribers[t]) + 1
	cs, ok := ps.subscribers[t]
	if !ok {
		cs = make(map[int]*Subscription)
		ps.subscribers[t] = cs
	}
	s := &Subscription{id: id, ps: ps, t: t, ch: make(chan Event, 10)}
	cs[id] = s

	return s
}

func (ps *PubSub) remove(t Topic, id int) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if chans, found := ps.subscribers[t]; found {
		delete(chans, id)
	}
}

// publish sends a message to all subscribed workers
func (ps *PubSub) publish(e Event) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if ss, found := ps.subscribers[e.Topic]; found {
		for _, s := range ss {
			s.send(e)
		}
	}
}

type Subscription struct {
	id   int
	ps   *PubSub
	t    Topic
	ch   chan Event
	once sync.Once
}

func (s *Subscription) Unsubscribe() {
	s.once.Do(func() {
		s.ps.remove(s.t, s.id)
		close(s.ch)
	})
}

func (s *Subscription) Consume() <-chan Event {
	return s.ch
}

func (s *Subscription) send(e Event) {
	select {
	case s.ch <- e:
	default:
		fmt.Println("Dropped message due to slow consumer:", e.Topic)
	}
}
