package event

import (
	"fmt"
	"sync"

	"github.com/alikarimi999/shahboard/pkg/log"
)

type Subscription interface {
	Topic() Topic        // Returns the topic associated with the subscription.
	Event() <-chan Event // Returns a channel to receive events.
	Err() <-chan error   // Returns a channel to receive errors.
	Unsubscribe()        // Unsubscribes from the topic and stops receiving events.
}

type EventHandler func(Event)

type SubscriptionManager struct {
	mu       sync.Mutex
	subs     map[string]Subscription
	newSubCh chan Subscription
	closeCh  chan struct{}
	wg       sync.WaitGroup
	l        log.Logger
	handler  EventHandler
}

func NewManager(l log.Logger, handler EventHandler) *SubscriptionManager {
	m := &SubscriptionManager{
		subs:     make(map[string]Subscription),
		newSubCh: make(chan Subscription),
		closeCh:  make(chan struct{}),
		l:        l,
		handler:  handler,
	}
	m.run()
	return m
}

func (m *SubscriptionManager) AddSubscription(sub Subscription) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.subs[sub.Topic().String()] = sub
	m.newSubCh <- sub
}

func (m *SubscriptionManager) RemoveSubscription(t Topic) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.subs, t.String())
}

func (m *SubscriptionManager) run() {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		for {
			select {
			case sub := <-m.newSubCh:
				m.wg.Add(1)
				go m.handleSubscription(sub)
			case <-m.closeCh:
				return
			}
		}
	}()
}

func (m *SubscriptionManager) handleSubscription(sub Subscription) {
	m.l.Debug(fmt.Sprintf("Listening to topic: '%s'", sub.Topic()))
	defer m.wg.Done()

	for {
		select {
		case e := <-sub.Event():
			m.handler(e)
		case err := <-sub.Err():
			m.l.Error(fmt.Sprintf("Error in subscription topic '%s': %v", sub.Topic(), err))
		case <-m.closeCh:
			return
		}
	}
}

func (m *SubscriptionManager) Stop() {
	close(m.closeCh)
	m.wg.Wait()
}
