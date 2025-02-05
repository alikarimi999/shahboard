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
	pongCh chan struct{}

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
		pongCh:      make(chan struct{}),
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
				s.l.Debug(fmt.Sprintf("session '%s' started listening to created games", s.id))
				defer wg.Done()
				for e := range s.sub.Event() {
					s.handleEvent(e)
				}
				s.l.Debug(fmt.Sprintf("session '%s' stopped listening to created games", s.id))
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
				Data: eve.Encode(),
			}

			s.sub.Unsubscribe()
			s.changeSub(s.em.SubscribeToGame(eve.GameID))

		case event.ActionEnded:
			msg = &Msg{
				MsgBase: MsgBase{
					Type:      MsgTypeGameEnd,
					Timestamp: time.Now().Unix(),
				},
				Data: e.Encode(),
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
				Data: e.Encode(),
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

	if s.sub == nil && (s.userId == data.User1.ID || s.userId == data.User2.ID) {
		s.changeSub(s.em.SubscribeToMatch(data.ID))
		s.role = gamePlayerRole
		s.l.Debug(fmt.Sprintf("session '%s' subscribed to match '%s'", s.id, data.ID))
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

		s.l.Debug(fmt.Sprintf("session '%s' subscribed to game '%s'", s.id, req.GameId))

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
		if err := s.p.Publish(event.EventGamePlayerMoved{
			ID:        req.ID,
			GameID:    req.GameID,
			PlayerID:  req.PlayerID,
			Move:      req.Move,
			Timestamp: req.Timestamp,
		}); err != nil {
			s.l.Error(fmt.Sprintf("failed to publish move event: %v", err))
		}
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
			ID:        types.NewObjectId(),
			Type:      MsgTypeWelcome,
			Timestamp: time.Now().Unix(),
		},
		Data: []byte("welcome"),
	})
}

func (s *session) sendPong() {
	s.pongCh <- struct{}{}
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
