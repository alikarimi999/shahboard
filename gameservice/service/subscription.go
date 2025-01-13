package game

import (
	"fmt"

	"sync"

	"github.com/alikarimi999/shahboard/event"
	"github.com/alikarimi999/shahboard/pkg/log"
)

type subscriptionManager struct {
	gs *Service

	mu   sync.Mutex
	subs map[string]event.Subscription

	newSubCh chan event.Subscription

	closeCh chan struct{}
	wg      sync.WaitGroup

	l log.Logger
}

func newSubscriptionManager(gs *Service) *subscriptionManager {
	sm := &subscriptionManager{
		gs:   gs,
		subs: make(map[string]event.Subscription),

		newSubCh: make(chan event.Subscription),
		closeCh:  make(chan struct{}),

		l: gs.l,
	}

	sm.run()

	return sm
}

func (m *subscriptionManager) addSub(sub event.Subscription) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.subs[sub.Topic().String()] = sub
	m.newSubCh <- sub
}

func (m *subscriptionManager) removeSub(sub event.Subscription) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.subs, sub.Topic().String())
}

func (m *subscriptionManager) run() {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		for {
			select {
			case sub := <-m.newSubCh:
				m.wg.Add(1)
				go func(sub event.Subscription) {
					m.l.Debug(fmt.Sprintf("listening to topic: '%s'", sub.Topic()))
					defer m.wg.Done()
					for {
						select {
						case e := <-sub.Event():
							m.gs.handleEvents(e)
						case err := <-sub.Err():
							fmt.Println(err)
						case <-m.closeCh:
							return
						}
					}
				}(sub)
			case <-m.closeCh:
				return
			}
		}
	}()
}

func (m *subscriptionManager) stop() {
	close(m.closeCh)
	m.wg.Wait()
}
