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

type gameRole uint8

const (
	gameViewerRole gameRole = iota + 1
	gamePlayerRole
)

type session struct {
	*websocket.Conn
	id     types.ObjectId
	userId types.ObjectId
	msgCh  chan []byte

	gameId types.ObjectId
	role   gameRole

	sub         event.Subscription
	changeSubCh chan event.Subscription

	em *gameEventsManager

	lastHeartBeat *atomic.Time

	closed *atomic.Bool
	stopCh chan struct{}

	p event.Publisher
	l log.Logger
}

func newSession(conn *websocket.Conn, em *gameEventsManager, userId types.ObjectId,
	p event.Publisher, l log.Logger) *session {
	sess := &session{
		Conn: conn,
		id:   types.NewObjectId(),

		userId:      userId,
		msgCh:       make(chan []byte, 100),
		changeSubCh: make(chan event.Subscription),
		em:          em,

		lastHeartBeat: atomic.NewTime(time.Now()),

		closed: atomic.NewBool(false),
		stopCh: make(chan struct{}),
		p:      p,
		l:      l,
	}
	sess.start()

	return sess
}

func (s *session) start() {
	go s.startListen()
}

func (s *session) startListen() {
	wg := sync.WaitGroup{}
	for {
		select {
		case <-s.stopCh:
			return
		case sub := <-s.changeSubCh:
			wg.Wait()
			s.sub = sub
			wg.Add(1)
			go func() {
				s.l.Debug(fmt.Sprintf("session '%d' started listening to '%s'", s.id, sub.Topic().String()))
				defer wg.Done()
				for e := range s.sub.Event() {
					s.handleEvent(e)
				}
				s.l.Debug(fmt.Sprintf("session '%d' stopped listening to '%s'", s.id, sub.Topic().String()))
			}()

		}
	}
}

func (s *session) changeSub(sub event.Subscription) {
	if sub == nil {
		return
	}
	s.changeSubCh <- sub
}

func (s *session) handleEvent(e event.Event) {

	var msg *Msg
	defer func() {
		if msg != nil {
			s.send(*msg)
		}
	}()

	switch e.GetTopic().Domain() {
	case event.DomainGame:
		switch e.GetAction() {
		case event.ActionCreated:
			eve := e.(*event.EventGameCreated)
			s.gameId = eve.GameID

			msg = &Msg{
				MsgBase: MsgBase{
					Type:      MsgTypeGameCreate,
					Timestamp: time.Now().Unix(),
				},
				Data: DataGameEvent{
					Domain: event.DomainGame,
					Action: event.ActionCreated.String(),
					Event:  eve.Encode(),
				}.Encode(),
			}

			s.sub.Unsubscribe()
			s.changeSub(s.em.SubscribeToGame(eve.GameID))

		case event.ActionEnded:
			msg = &Msg{
				MsgBase: MsgBase{
					Type:      MsgTypeGameEnd,
					Timestamp: time.Now().Unix(),
				},
				Data: DataGameEvent{
					Domain: event.DomainGame,
					Action: event.ActionEnded.String(),
					Event:  e.Encode(),
				}.Encode(),
			}
			s.sub.Unsubscribe()
			s.sub = nil

		default:
			var mt MsgType
			switch e.GetAction() {
			case event.ActionCreated:
				mt = MsgTypeGameCreate
			case event.ActionGameMoveApprove:
				mt = MsgTypeMoveApproved
			default:
				return
			}
			msg = &Msg{
				MsgBase: MsgBase{
					Type:      mt,
					Timestamp: time.Now().Unix(),
				},
				Data: DataGameEvent{
					Domain: event.DomainGame,
					Action: e.GetAction().String(),
					Event:  e.Encode(),
				}.Encode(),
			}
		}
	}
}

func (s *session) handleFindMatchRequest(msgId types.ObjectId, data DataFindMatchRequest) {
	var errMsg string
	defer func() {
		if errMsg != "" {
			s.sendErr(msgId, errMsg)
		}
	}()

	if s.sub == nil && (s.userId == data.User1 || s.userId == data.User2) {
		s.changeSub(s.em.SubscribeToMatch(data.MatchID))
		s.role = gamePlayerRole
		return
	}

	errMsg = "already subscribed to a game"
}

func (s *session) handleViewGameRequest(msgId types.ObjectId, req DataGameViewRequest) {
	var errMsg string
	var msg *Msg
	defer func() {
		if msg != nil {
			s.send(*msg)
		}
		if errMsg != "" {
			s.sendErr(msgId, errMsg)
		}
	}()

	if s.sub == nil {
		msg = &Msg{
			MsgBase: MsgBase{
				ID:        msgId,
				Type:      MsgTypeView,
				Timestamp: time.Now().Unix(),
			}}

		s.changeSub(s.em.SubscribeToGame(req.GameId))
		s.role = gameViewerRole

		s.l.Debug(fmt.Sprintf("session '%d' subscribed to game '%d'", s.id, req.GameId))

		return
	}

	errMsg = "already subscribed to a game"
}

func (s *session) handleMoveRequest(msgId types.ObjectId, req DataGamePlayerMoveRequest) {
	var errMsg string
	defer func() {
		if errMsg != "" {
			s.sendErr(msgId, errMsg)
		}
	}()

	if s.sub != nil && s.role == gamePlayerRole && s.userId == req.PlayerID && s.gameId == req.GameID {
		s.l.Debug(fmt.Sprintf("session '%d' received move request %d", s.id, req.ID))
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

func (s *session) send(msg Msg) {
	b, err := json.Marshal(msg)
	if err != nil {
		s.l.Error(fmt.Sprintf("failed to marshal message: %v", err))
		return
	}
	s.msgCh <- b
}

func (s *session) sendErr(id types.ObjectId, err string) {
	s.send(Msg{
		MsgBase: MsgBase{
			ID:        id,
			Type:      MsgTypeError,
			Timestamp: time.Now().Unix(),
		},
		Data: []byte(err),
	})
}

func (s *session) sendWelcome() {
	s.send(Msg{
		MsgBase: MsgBase{
			ID:        0,
			Type:      MsgTypeWelcome,
			Timestamp: time.Now().Unix(),
		},
		Data: []byte("welcome"),
	})
}

func (s *session) isSubscribedToGame() bool {
	return s.sub != nil
}

func (s *session) close() error {
	s.closed.Store(true)
	close(s.stopCh)
	return s.Conn.Close()
}

func (s *session) isClosed() bool {
	return s.closed.Load()
}

func (s *session) reconnect(conn *websocket.Conn) {
	s.closed.Store(false)
	s.stopCh = make(chan struct{})
	s.Conn = conn
}
