package ws

import (
	"sync"
	"time"

	"github.com/alikarimi999/shahboard/event"
	"github.com/alikarimi999/shahboard/pkg/log"
	"github.com/alikarimi999/shahboard/types"
)

var (
	defaultMatchAndGameExpireTime = 5 * time.Minute
	defaultBroadcastInterval      = 1 * time.Second
)

type sessionsEventsHandler struct {
	directChatSub event.Subscription

	gameSub     event.Subscription
	gameChatSub event.Subscription

	em *endedGamesList

	gmu               sync.RWMutex
	createdGameEvents map[types.ObjectId]event.Event // map by matchId

	broadcastTicker *time.Ticker
	cleanupTicker   *time.Ticker

	dmu                    sync.Mutex
	direchtChatSubSessions map[types.ObjectId]*session // map by userId

	mmu              sync.Mutex
	matchSubSessions map[types.ObjectId][]struct {
		*session
		addedAt time.Time
	}

	gcmu                    sync.RWMutex
	gameWithChatSubSessions map[types.ObjectId]*gameSubscribers

	gameExpireTime time.Duration
	l              log.Logger
	stopCh         chan struct{}
}

func newSessionsEventsHandler(s event.Subscriber, em *endedGamesList, l log.Logger) *sessionsEventsHandler {
	m := &sessionsEventsHandler{
		gameSub:       s.Subscribe(event.TopicGame),
		gameChatSub:   s.Subscribe(event.TopicGameChat),
		em:            em,
		directChatSub: s.Subscribe(event.TopicDirectChat),

		gameExpireTime:    defaultMatchAndGameExpireTime,
		createdGameEvents: make(map[types.ObjectId]event.Event),

		broadcastTicker: time.NewTicker(defaultBroadcastInterval),
		cleanupTicker:   time.NewTicker(defaultMatchAndGameExpireTime),

		direchtChatSubSessions: make(map[types.ObjectId]*session),
		matchSubSessions: make(map[types.ObjectId][]struct {
			*session
			addedAt time.Time
		}),
		gameWithChatSubSessions: make(map[types.ObjectId]*gameSubscribers),

		l:      l,
		stopCh: make(chan struct{}),
	}

	m.run()

	return m
}

func (m *sessionsEventsHandler) run() {
	m.startEventListener()
	m.startCleanupRoutine()
	m.startBroadcastRoutine()
}

func (h *sessionsEventsHandler) startEventListener() {
	go func() {
		for {
			select {
			case <-h.stopCh:
				h.l.Debug("event listener stopped")
				return
			case e := <-h.gameSub.Event():
				switch e.GetTopic().Action() {
				case event.ActionCreated:
					eve, ok := e.(*event.EventGameCreated)
					if !ok {
						h.l.Warn("invalid event type for ActionCreated")
						continue
					}

					// broadcast to all sessisons
					h.mmu.Lock()
					for _, s := range h.matchSubSessions[eve.MatchID] {
						s.consume(e)
					}
					delete(h.matchSubSessions, eve.MatchID)
					h.mmu.Unlock()

					// store created game events for situations where the user websocket connection
					// request for the evnet after the event is received
					h.gmu.Lock()
					h.createdGameEvents[eve.MatchID] = e
					h.gmu.Unlock()
				default:

					gameID, err := types.ParseObjectId(e.GetTopic().Resource())
					if err != nil {
						continue
					}

					if e.GetTopic().Action() == event.ActionEnded {
						h.em.add(gameID)
					}

					var gs *gameSubscribers
					var ok bool
					h.gcmu.RLock()
					gs, ok = h.gameWithChatSubSessions[gameID]
					h.gcmu.RUnlock()

					if ok {
						gs.sendEvent(e)
					}
				}

			case e := <-h.gameChatSub.Event():
				gameID, err := types.ParseObjectId(e.GetTopic().Resource())
				if err != nil {
					continue
				}

				var gs *gameSubscribers
				var ok bool
				h.gcmu.RLock()
				gs, ok = h.gameWithChatSubSessions[gameID]
				h.gcmu.RUnlock()

				if ok {
					gs.sendEvent(e)
				}
			}
		}
	}()
}

func (h *sessionsEventsHandler) startCleanupRoutine() {
	go func() {
		for {
			select {
			case <-h.stopCh:
				h.l.Debug("cleanup routine stopped")
				return
			case <-h.cleanupTicker.C:
				h.gmu.Lock()
				for k, v := range h.createdGameEvents {
					if time.Since(time.Unix(v.TimeStamp(), 0)) > h.gameExpireTime {
						delete(h.createdGameEvents, k)
					}
				}
				h.gmu.Unlock()

				h.mmu.Lock()
				for key, ss := range h.matchSubSessions {
					newSessions := make([]struct {
						*session
						addedAt time.Time
					}, 0, len(ss))
					for _, s := range ss {
						if time.Since(s.addedAt) < h.gameExpireTime {
							newSessions = append(newSessions, s)
						}
					}

					if len(newSessions) == 0 {
						delete(h.matchSubSessions, key)
					} else {
						h.matchSubSessions[key] = newSessions
					}
				}
				h.mmu.Unlock()
			}
		}
	}()
}

func (h *sessionsEventsHandler) startBroadcastRoutine() {
	go func() {
		for {
			select {
			case <-h.stopCh:
				h.l.Debug("broadcast routine stopped")
				return
			case <-h.broadcastTicker.C:
				h.mmu.Lock()
				h.gmu.RLock()
				for matchId, e := range h.createdGameEvents {
					for _, s := range h.matchSubSessions[matchId] {
						s.consume(e)
					}
					delete(h.createdGameEvents, matchId)
				}
				h.gmu.RUnlock()
				h.mmu.Unlock()
			}
		}
	}()
}

func (m *sessionsEventsHandler) stop() {
	close(m.stopCh)
	m.broadcastTicker.Stop()
	m.cleanupTicker.Stop()
}

func (h *sessionsEventsHandler) subscribeToBasicEvents(s *session) {
	h.dmu.Lock()
	defer h.dmu.Unlock()
	h.direchtChatSubSessions[s.userId] = s
}

func (h *sessionsEventsHandler) subscribeToMatch(s *session) {
	h.mmu.Lock()
	defer h.mmu.Unlock()
	h.matchSubSessions[s.matchId.Load()] = append(h.matchSubSessions[s.matchId.Load()],
		struct {
			*session
			addedAt time.Time
		}{
			session: s,
			addedAt: time.Now(),
		})
}

func (m *sessionsEventsHandler) subscribeToGameWithChat(s *session, gameId types.ObjectId) {
	m.gcmu.Lock()
	defer m.gcmu.Unlock()

	if gs, ok := m.gameWithChatSubSessions[gameId]; ok {
		gs.add(s)
	} else {
		m.gameWithChatSubSessions[gameId] = newGameSubscribers()
		m.gameWithChatSubSessions[gameId].add(s)
	}
}

func (m *sessionsEventsHandler) unsubscribeFromBasicEvents(s *session) {
	m.dmu.Lock()
	defer m.dmu.Unlock()
	delete(m.direchtChatSubSessions, s.userId)
}

func (m *sessionsEventsHandler) unsubscribeFromMatch(s *session) {
	m.mmu.Lock()
	defer m.mmu.Unlock()

	sessions := m.matchSubSessions[s.matchId.Load()]
	newSessions := make([]struct {
		*session
		addedAt time.Time
	}, 0, len(sessions))

	for _, ses := range sessions {
		if ses.id != s.id {
			newSessions = append(newSessions, ses)
		}
	}

	if len(newSessions) == 0 {
		delete(m.matchSubSessions, s.matchId.Load())
	} else {
		m.matchSubSessions[s.matchId.Load()] = newSessions
	}
}

func (m *sessionsEventsHandler) unsubscribeFromGameWithChat(s *session, gamesId ...types.ObjectId) {
	// not sure if RLock will cause erroro or not
	m.gcmu.RLock()
	defer m.gcmu.RUnlock()

	for _, gameId := range gamesId {
		if ss, ok := m.gameWithChatSubSessions[gameId]; ok {
			ss.remove(s)
		}
	}
}

func (m *sessionsEventsHandler) deleteGameSubscribers(gamesId ...types.ObjectId) {
	m.gcmu.Lock()
	defer m.gcmu.Unlock()
	for _, gameId := range gamesId {
		delete(m.gameWithChatSubSessions, gameId)
	}
}

type gameSubscribers struct {
	sync.RWMutex
	subscribers map[types.ObjectId]*session // map by sessionId
}

func newGameSubscribers() *gameSubscribers {
	return &gameSubscribers{
		subscribers: make(map[types.ObjectId]*session),
	}
}

func (g *gameSubscribers) add(s *session) {
	g.Lock()
	defer g.Unlock()
	g.subscribers[s.id] = s
}

func (g *gameSubscribers) remove(s *session) {
	g.Lock()
	defer g.Unlock()
	delete(g.subscribers, s.id)
}

func (g *gameSubscribers) sendEvent(e event.Event) {
	g.RLock()
	defer g.RUnlock()

	for _, s := range g.subscribers {
		s.consume(e)
	}
}
