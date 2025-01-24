package ws

import (
	"encoding/json"
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

type session struct {
	*websocket.Conn
	id     types.ObjectId
	userId types.ObjectId
	msgCh  chan []byte

	currentGame     gameSubscription
	matchRequesteId *atomic.Uint64

	lastHeartBeat *atomic.Time

	s  event.Subscriber
	p  event.Publisher
	sm *event.SubscriptionManager

	closed *atomic.Bool
	stopCh chan struct{}
	pongCh chan string
}

func newSession(conn *websocket.Conn, userId types.ObjectId,
	s event.Subscriber, p event.Publisher, l log.Logger) *session {
	sess := &session{
		Conn: conn,
		id:   types.NewObjectId(),

		userId: userId,
		msgCh:  make(chan []byte, 100),

		matchRequesteId: atomic.NewUint64(0),

		lastHeartBeat: atomic.NewTime(time.Now()),

		s: s,
		p: p,

		closed: atomic.NewBool(false),
		stopCh: make(chan struct{}),
		pongCh: make(chan string),
	}
	sess.sm = event.NewManager(l, sess.eventHandler())
	// sess.sm.AddSubscription(s.Subscribe(event.TopicGame))
	// sess.sm.AddSubscription(s.Subscribe(event.TopicUsersMatched))

	return sess
}

func (s *session) eventHandler() event.EventHandler {
	return func(e event.Event) {

		var msgId types.ObjectId
		switch e.GetTopic().Domain() {
		case event.DomainGame:
			switch e.GetAction() {
			case event.ActionCreated:
				eve := e.(*event.EventGameCreated)
				if s.matchRequesteId.Load() != 0 && (s.userId == eve.Player1.ID || s.userId == eve.Player2.ID) {
					s.currentGame = gameSubscription{
						requestID: types.ObjectId(s.matchRequesteId.Load()),
						gameID:    eve.GameID,
						role:      gamePlayerRole,
					}
					msgId = types.ObjectId(s.matchRequesteId.Load())
					s.matchRequesteId.Store(0)
					break
				}

				return
			case event.ActionEnded:
				if e.GetTopic().Resource() == s.currentGame.gameID.String() {
					msgId = s.currentGame.requestID
					s.currentGame = gameSubscription{}
					s.sm.RemoveSubscription(event.TopicGame)
					break
				}

				return
			default:
				if e.GetTopic().Resource() == s.currentGame.gameID.String() {
					msgId = s.currentGame.requestID
					break
				}

				return
			}

		case event.DomainMatch:
			switch e.GetAction() {
			case event.ActionPlayersMatched:
				if s.matchRequesteId.Load() != 0 {
					eve := e.(*event.EventUsersMatched)
					if s.userId == eve.User1.ID || s.userId == eve.User2.ID {
						msgId = types.ObjectId(s.matchRequesteId.Load())
						s.sm.RemoveSubscription(event.TopicUsersMatched)
						break
					}
				}

				return
			default:
				return
			}
		default:
			return
		}

		s.send(serverMsg{
			msgBase: msgBase{
				ID:        msgId,
				Type:      msgTypeView,
				Timestamp: time.Now().Unix(),
			},
			Data: e.Encode(),
		})
	}
}

func (s *session) send(msg serverMsg) {
	b, _ := json.Marshal(msg)
	s.msgCh <- b
}

func (s *session) viewGame(msgId, gameId types.ObjectId) {
	if s.currentGame.gameID != 0 {
		s.currentGame = gameSubscription{
			requestID: msgId,
			gameID:    gameId,
			role:      gameViewerRole,
		}

		s.send(serverMsg{
			msgBase: msgBase{
				ID:        msgId,
				Type:      msgTypeView,
				Timestamp: time.Now().Unix(),
			},
		})
		return
	}
	s.sendErr(msgId, "already subscribed to a game")
}

func (s *session) gameSubscribed() types.ObjectId {
	return s.currentGame.gameID
}

func (s *session) gameUnsubscribe() {
	s.currentGame = gameSubscription{}
}

func (s *session) isPlaying() bool {
	return s.currentGame.role == gamePlayerRole
}

func (s *session) sendErr(id types.ObjectId, err string) {
	s.send(serverMsg{
		msgBase: msgBase{
			ID:        id,
			Type:      msgTypeErr,
			Timestamp: time.Now().Unix(),
		},
		Data: []byte(err),
	})
}

func (s *session) sendWelcome() {
	s.send(serverMsg{
		msgBase: msgBase{
			ID:        0,
			Type:      msgTypeWelcome,
			Timestamp: time.Now().Unix(),
		},
		Data: []byte("welcome"),
	})
}
