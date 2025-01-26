package ws

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/alikarimi999/shahboard/event"
	"github.com/alikarimi999/shahboard/pkg/log"
	"github.com/alikarimi999/shahboard/types"
	"github.com/gorilla/websocket"
	"go.uber.org/atomic"
)

const (
	gameViewerRole uint8 = iota + 1
	gamePlayerRole
)

type gameSubscription struct {
	requestID types.ObjectId
	gameID    types.ObjectId
	role      uint8
}

type matchRequest struct {
	matchId types.ObjectId
	msgId   types.ObjectId
}

type session struct {
	*websocket.Conn
	id     types.ObjectId
	userId types.ObjectId
	msgCh  chan []byte

	mu                  sync.Mutex
	subscribedTopicGame event.Topic
	currentGame         *gameSubscription
	matchReq            *matchRequest

	lastHeartBeat *atomic.Time

	s  event.Subscriber
	p  event.Publisher
	sm *event.SubscriptionManager

	closed *atomic.Bool
	stopCh chan struct{}
	pongCh chan string

	l log.Logger
}

func newSession(conn *websocket.Conn, userId types.ObjectId,
	s event.Subscriber, p event.Publisher, l log.Logger) *session {
	sess := &session{
		Conn: conn,
		id:   types.NewObjectId(),

		userId: userId,
		msgCh:  make(chan []byte, 100),

		lastHeartBeat: atomic.NewTime(time.Now()),

		s: s,
		p: p,

		closed: atomic.NewBool(false),
		stopCh: make(chan struct{}),
		pongCh: make(chan string),
		l:      l,
	}
	sess.sm = event.NewManager(l, sess.eventHandler())

	return sess
}

func (s *session) eventHandler() event.EventHandler {
	return func(e event.Event) {

		var msg *ServerMsg
		s.mu.Lock()
		defer func() {
			// first unlock then send message to avoid deadlock
			s.mu.Unlock()
			if msg != nil {
				s.send(*msg)
			}
		}()

		switch e.GetTopic().Domain() {
		case event.DomainGame:
			switch e.GetAction() {
			case event.ActionCreated:
				eve := e.(*event.EventGameCreated)
				if s.matchReq != nil && s.matchReq.matchId == eve.MatchID {
					s.currentGame = &gameSubscription{
						requestID: s.matchReq.msgId,
						gameID:    eve.GameID,
						role:      gamePlayerRole,
					}
					msg = &ServerMsg{
						MsgBase: MsgBase{
							ID:        s.matchReq.msgId,
							Type:      MsgTypeEvent,
							Timestamp: time.Now().Unix(),
						},
						Data: DataGameEvent{
							Domain: event.DomainGame,
							Action: event.ActionCreated.String(),
							Event:  eve,
						},
					}

					s.matchReq = nil
					s.sm.RemoveSubscription(s.subscribedTopicGame)
					sub := s.s.Subscribe(event.TopicGame.WithResource(eve.GameID.String()))
					s.subscribedTopicGame = sub.Topic()
					s.sm.AddSubscription(sub)
				}

			case event.ActionEnded:
				s.currentGame = nil
				s.sm.RemoveSubscription(s.subscribedTopicGame)
				msg = &ServerMsg{
					MsgBase: MsgBase{
						ID:        s.currentGame.requestID,
						Type:      MsgTypeEvent,
						Timestamp: time.Now().Unix(),
					},
					Data: DataGameEvent{
						Domain: event.DomainGame,
						Action: event.ActionEnded.String(),
						Event:  e,
					},
				}
			default:
				msg = &ServerMsg{
					MsgBase: MsgBase{
						ID:        s.currentGame.requestID,
						Type:      MsgTypeEvent,
						Timestamp: time.Now().Unix(),
					},
					Data: DataGameEvent{
						Domain: event.DomainGame,
						Action: e.GetAction().String(),
						Event:  e,
					},
				}

			}

		}
	}
}

func (s *session) handlerFindMatchRequest(msgId types.ObjectId, data DataFindMatchRequest) {
	var errMsg string
	s.mu.Lock()
	defer func() {
		s.mu.Unlock()
		if errMsg != "" {
			s.sendErr(msgId, errMsg)
		}
	}()

	if s.matchReq == nil && s.currentGame == nil && (s.userId == data.User1 || s.userId == data.User2) {

		s.matchReq = &matchRequest{
			matchId: data.MatchID,
			msgId:   msgId,
		}

		sub := s.s.Subscribe(event.TopicGame)
		s.subscribedTopicGame = sub.Topic()
		s.sm.AddSubscription(sub)

		return
	}

	errMsg = "already subscribed to a game"
}

func (s *session) handleViewGameRequest(msgId types.ObjectId, req DataGameViewRequest) {
	var errMsg string
	var msg *ServerMsg
	s.mu.Lock()
	defer func() {
		s.mu.Unlock()
		if msg != nil {
			s.send(*msg)
		}
		if errMsg != "" {
			s.sendErr(msgId, errMsg)
		}
	}()

	if s.currentGame == nil && s.matchReq == nil {
		s.currentGame = &gameSubscription{
			requestID: msgId,
			gameID:    req.GameId,
			role:      gameViewerRole,
		}

		msg = &ServerMsg{
			MsgBase: MsgBase{
				ID:        msgId,
				Type:      MsgTypeView,
				Timestamp: time.Now().Unix(),
			}}

		sub := s.s.Subscribe(event.TopicGame.WithResource(req.GameId.String()))
		s.subscribedTopicGame = sub.Topic()
		s.sm.AddSubscription(sub)

		return
	}

	errMsg = "already subscribed to a game"
}

func (s *session) handleMoveRequest(msgId types.ObjectId, req DataGamePlayerMoveRequest) {
	var errMsg string
	s.mu.Lock()
	defer func() {
		s.mu.Unlock()
		if errMsg != "" {
			s.sendErr(msgId, errMsg)
		}
	}()

	if s.currentGame != nil && s.currentGame.role == gamePlayerRole && s.currentGame.gameID == req.GameID {
		s.p.Publish(event.EventGamePlayerMoved{
			ID:        req.ID,
			GameID:    req.GameID,
			PlayerID:  req.PlayerID,
			Move:      req.Move,
			Timestamp: req.Timestamp,
		})

		return
	}

	errMsg = "not allowd"
}

func (s *session) send(msg ServerMsg) {
	b, err := json.Marshal(msg)
	if err != nil {
		s.l.Error(fmt.Sprintf("failed to marshal message: %v", err))
		return
	}
	s.msgCh <- b
}

func (s *session) sendErr(id types.ObjectId, err string) {
	s.send(ServerMsg{
		MsgBase: MsgBase{
			ID:        id,
			Type:      MsgTypeError,
			Timestamp: time.Now().Unix(),
		},
		Data: DataError(err),
	})
}

func (s *session) sendWelcome() {
	s.send(ServerMsg{
		MsgBase: MsgBase{
			ID:        0,
			Type:      MsgTypeWelcome,
			Timestamp: time.Now().Unix(),
		},
		Data: DataWelcodme("welcome"),
	})
}

func (s *session) isSubscribedToGame() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.currentGame != nil || s.matchReq != nil
}

func (s *session) close() error {
	s.closed.Store(true)
	close(s.stopCh)
	return s.Conn.Close()
}

func (s *session) isClosed() bool {
	return s.closed.Load()
}

func (s *session) reconnect(conn *websocket.Conn) bool {
	if s.isClosed() {
		s.closed.Store(false)
		s.stopCh = make(chan struct{})
		s.Conn = conn
		return true
	}

	return false
}
