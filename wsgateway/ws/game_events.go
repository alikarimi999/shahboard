package ws

import (
	"sync"
	"time"

	"github.com/alikarimi999/shahboard/event"
	"github.com/alikarimi999/shahboard/pkg/log"
	"github.com/alikarimi999/shahboard/types"
)

var (
	defaultGameExpireTime    = 5 * time.Minute
	defaultBroadcastInterval = 1 * time.Second
)

type gameEventsManager struct {
	gameSub event.Subscription
	chatSub event.Subscription

	mu                sync.Mutex
	createdGameEvents map[types.ObjectId]event.Event // map by match id

	broadcastTicker *time.Ticker
	cleanupTicker   *time.Ticker

	smu              sync.Mutex
	matchSubs        map[types.ObjectId][]*subscription
	gameWithChatSubs map[types.ObjectId][]*subscription

	gameExpireTime time.Duration
	l              log.Logger
	stopCh         chan struct{}
}

func newGameEventsManager(s event.Subscriber, l log.Logger) *gameEventsManager {
	m := &gameEventsManager{
		gameSub:           s.Subscribe(event.TopicGame),
		chatSub:           s.Subscribe(event.TopicGameChat),
		gameExpireTime:    defaultGameExpireTime,
		createdGameEvents: make(map[types.ObjectId]event.Event),

		broadcastTicker: time.NewTicker(defaultBroadcastInterval),
		cleanupTicker:   time.NewTicker(defaultGameExpireTime),

		matchSubs:        make(map[types.ObjectId][]*subscription),
		gameWithChatSubs: make(map[types.ObjectId][]*subscription),
		l:                l,
		stopCh:           make(chan struct{}),
	}

	m.run()

	return m
}

func (m *gameEventsManager) run() {
	m.startEventListener()
	m.startCleanupRoutine()
	m.startBroadcastRoutine()
}

func (m *gameEventsManager) startEventListener() {
	go func() {
		for {
			select {
			case <-m.stopCh:
				m.l.Debug("event listener stopped")
				return
			case e := <-m.gameSub.Event():
				switch e.GetAction() {
				case event.ActionCreated:
					eve, ok := e.(*event.EventGameCreated)
					if !ok {
						m.l.Warn("invalid event type for ActionCreated")
						continue
					}

					// broadcast to all subscribers
					subs, ok := m.matchSubs[eve.MatchID]
					if ok {
						for _, sub := range subs {
							sub.ch <- e
						}
					}

					// store created game events for situations where the user websocket connection
					// established after receiving the event
					m.mu.Lock()
					m.createdGameEvents[eve.MatchID] = e
					m.mu.Unlock()
				default:
					gameID, err := types.ParseObjectId(e.GetTopic().Resource())
					if err != nil {
						continue
					}

					// other game events only broadcast to subscribers that subscribed before receiving the event
					m.smu.Lock()
					for _, subs := range m.gameWithChatSubs[gameID] {
						select {
						case subs.ch <- e:
						default:
							m.l.Warn("failed to broadcast event to subscriber")
						}
					}
					m.smu.Unlock()
				}

			case e := <-m.chatSub.Event():
				gameID, err := types.ParseObjectId(e.GetTopic().Resource())
				if err != nil {
					continue
				}

				m.smu.Lock()
				for _, subs := range m.gameWithChatSubs[gameID] {
					select {
					case subs.ch <- e:
					default:
						m.l.Warn("failed to broadcast event to subscriber")
					}
				}
				m.smu.Unlock()

			}
		}
	}()
}

func (m *gameEventsManager) startCleanupRoutine() {
	go func() {
		for {
			select {
			case <-m.stopCh:
				m.l.Debug("cleanup routine stopped")
				return
			case <-m.cleanupTicker.C:
				m.mu.Lock()
				for k, v := range m.createdGameEvents {
					if time.Since(time.Unix(v.TimeStamp(), 0)) > m.gameExpireTime {
						delete(m.createdGameEvents, k)
					}
				}
				m.mu.Unlock()
			}
		}
	}()
}

func (m *gameEventsManager) startBroadcastRoutine() {
	go func() {
		for {
			select {
			case <-m.stopCh:
				m.l.Debug("broadcast routine stopped")
				return
			case <-m.broadcastTicker.C:
				m.smu.Lock()
				m.mu.Lock()
				for id, e := range m.createdGameEvents {
					subs, ok := m.matchSubs[id]
					if ok {
						for _, sub := range subs {
							select {
							case sub.ch <- e:
							default:
								m.l.Warn("failed to broadcast event to subscriber")
							}
						}
					}
				}
				m.mu.Unlock()
				m.smu.Unlock()
			}
		}
	}()
}

func (m *gameEventsManager) Stop() {
	close(m.stopCh)
	m.broadcastTicker.Stop()
	m.cleanupTicker.Stop()

	m.smu.Lock()
	defer m.smu.Unlock()

	for _, subs := range m.matchSubs {
		for _, sub := range subs {
			close(sub.ch)
			close(sub.errCh)
		}
	}

	for _, subs := range m.gameWithChatSubs {
		for _, sub := range subs {
			close(sub.ch)
			close(sub.errCh)
		}
	}
}

func (m *gameEventsManager) subscribeToMatch(matchId types.ObjectId) event.Subscription {
	m.smu.Lock()
	defer m.smu.Unlock()

	s := &subscription{
		index:   len(m.matchSubs[matchId]),
		matchId: matchId,
		m:       m,
		topic:   event.TopicUsersMatched,
		ch:      make(chan event.Event, 100),
		errCh:   make(chan error),
	}

	m.matchSubs[matchId] = append(m.matchSubs[matchId], s)
	return s
}

func (m *gameEventsManager) subscribeToGameWithChat(gameId types.ObjectId) event.Subscription {
	m.smu.Lock()
	defer m.smu.Unlock()

	s := &subscription{
		index:  len(m.gameWithChatSubs[gameId]),
		gameId: gameId,
		m:      m,
		topic:  event.TopicGame.WithResource(gameId.String()),
		ch:     make(chan event.Event, 100),
		errCh:  make(chan error),
	}

	m.gameWithChatSubs[gameId] = append(m.gameWithChatSubs[gameId], s)
	return s
}

func (m *gameEventsManager) removeSubscription(sub *subscription) {
	m.smu.Lock()
	defer m.smu.Unlock()

	if !sub.matchId.IsZero() {
		if subs, ok := m.matchSubs[sub.matchId]; ok {
			for i, s := range subs {
				if s.index == sub.index {
					m.matchSubs[sub.matchId] = append(subs[:i], subs[i+1:]...)
					return
				}
			}
		}
	}

	if !sub.gameId.IsZero() {
		if subs, ok := m.gameWithChatSubs[sub.gameId]; ok {
			for i, s := range subs {
				if s.index == sub.index {
					m.gameWithChatSubs[sub.gameId] = append(subs[:i], subs[i+1:]...)
					return
				}
			}
		}
	}
}

// implementaion of event.Subscription
type subscription struct {
	index           int
	matchId, gameId types.ObjectId
	m               *gameEventsManager
	topic           event.Topic
	ch              chan event.Event
	errCh           chan error
}

func (s *subscription) Topic() event.Topic {
	return s.topic
}

func (s *subscription) Event() <-chan event.Event {
	return s.ch
}

func (s *subscription) Err() <-chan error {
	return s.errCh
}

func (s *subscription) Unsubscribe() {
	close(s.ch)
	close(s.errCh)
	s.m.removeSubscription(s)
}
