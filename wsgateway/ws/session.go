package ws

import (
	"context"
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

	rc *redisCache

	em *gameEventsManager

	lastHeartBeat *atomic.Time

	closed *atomic.Bool
	stopCh chan struct{}

	p event.Publisher
	l log.Logger
}

func newSession(conn *websocket.Conn, em *gameEventsManager, rc *redisCache, userId types.ObjectId,
	gameId types.ObjectId, p event.Publisher, l log.Logger) *session {
	sess := &session{
		Conn: conn,
		id:   types.NewObjectId(),

		userId:      userId,
		gameId:      gameId,
		msgCh:       make(chan []byte, 100),
		pongCh:      make(chan struct{}),
		changeSubCh: make(chan event.Subscription),
		rc:          rc,
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
				defer wg.Done()
				t := s.sub.Topic().String()
				s.l.Debug(fmt.Sprintf("session '%s' subscribed to '%s'", s.id, t))
				for e := range s.sub.Event() {
					s.handleEvent(e)
				}
				s.l.Debug(fmt.Sprintf("session '%s' unsubscribed from '%s'", s.id, t))
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
			if s.isClosed() {
				if err := s.rc.SaveSessionMsg(context.Background(), s.id, msg); err != nil {
					s.l.Error(fmt.Sprintf("session '%s' failed to save message to redis: %v", s.id, err))
					return
				}
				s.l.Debug(fmt.Sprintf("session '%s' saved message to redis", s.id))
				return
			}
			s.send(*msg)
		}
	}()

	switch e.GetTopic().Domain() {
	case event.DomainGame:
		msg = s.handleGameEvent(e)
	case event.DomainGameChat:
		msg = s.handleGameChatEvent(e)
	}
}

func (s *session) handleGameEvent(e event.Event) *Msg {
	var msg *Msg

	switch e.GetAction() {
	case event.ActionCreated:
		eve := e.(*event.EventGameCreated)
		if eve.Player1.ID != s.userId && eve.Player2.ID != s.userId {
			s.sub.Unsubscribe()
			return nil
		}

		s.gameId = eve.GameID

		msg = &Msg{
			MsgBase: MsgBase{
				Type:      MsgTypeGameCreate,
				Timestamp: time.Now().Unix(),
			},
			Data: eve.Encode(),
		}

		s.sub.Unsubscribe()
		s.changeSub(s.em.subscribeToGameWithChat(eve.GameID))

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
		s.gameId = types.ObjectZero
		s.role = 0

	default:
		var mt MsgType
		switch e.GetAction() {
		case event.ActionCreated:
			mt = MsgTypeGameCreate
		case event.ActionGameMoveApprove:
			mt = MsgTypeMoveApproved
		case event.ActionGamePlayerConnectionUpdated:
			mt = MsgTypePlayerConnectionUpdated
		default:
			return nil
		}
		msg = &Msg{
			MsgBase: MsgBase{
				Type:      mt,
				Timestamp: time.Now().Unix(),
			},
			Data: e.Encode(),
		}
	}

	return msg
}

func (s *session) handleGameChatEvent(e event.Event) *Msg {
	var mt MsgType
	switch e.GetAction() {
	case event.ActionCreated:
		mt = MsgTypeChatCreated
	case event.ActionMsgApproved:
		mt = MsgTypeChatMsgApproved
	}

	if mt != "" {
		return &Msg{
			MsgBase: MsgBase{
				Type:      mt,
				Timestamp: time.Now().Unix(),
			},
			Data: e.Encode(),
		}
	}

	return nil
}

// func (s *session) subscribeAsPlayer(gameId types.ObjectId) {
// 	s.changeSub(s.em.subscribeToGameWithChat(gameId))
// 	s.gameId = gameId
// 	s.role = gamePlayerRole
// 	s.l.Debug(fmt.Sprintf("session '%s' subscribed to game '%s'", s.id, gameId))
// }

// func (s *session) subscribeAsViewer(gameId types.ObjectId) {
// 	s.changeSub(s.em.subscribeToGameWithChat(gameId))
// 	s.gameId = gameId
// 	s.role = gameViewerRole
// 	s.l.Debug(fmt.Sprintf("session '%s' subscribed to game '%s'", s.id, gameId))
// }

func (s *session) handleFindMatchRequest(msgId types.ObjectId, data dataFindMatchRequest) {
	var errMsg string
	defer func() {
		if errMsg != "" {
			s.sendErr(msgId, errMsg)
		}
	}()

	if s.sub == nil && (s.userId == data.User1.ID || s.userId == data.User2.ID) {
		s.changeSub(s.em.subscribeToMatch(data.ID))
		s.role = gamePlayerRole
		return
	}

	errMsg = "already subscribed to a game"
}

func (s *session) handleResumeGameRequest(msgId types.ObjectId, req dataResumeGameRequest) {
	var errMsg string
	defer func() {
		if errMsg != "" {
			s.sendErr(msgId, errMsg)
		}
	}()

	if s.sub == nil && s.gameId == req.GameId {
		s.changeSub(s.em.subscribeToGameWithChat(req.GameId))
		s.role = gamePlayerRole
		s.l.Debug(fmt.Sprintf("session '%s' subscribed to game '%s'", s.id, req.GameId))
		return
	}

	errMsg = "already subscribed to a game"
}

func (s *session) handleViewGameRequest(msgId types.ObjectId, req dataGameViewRequest) {
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

		s.changeSub(s.em.subscribeToGameWithChat(req.GameId))
		s.role = gameViewerRole

		s.l.Debug(fmt.Sprintf("session '%s' subscribed to game '%s'", s.id, req.GameId))

		return
	}

	errMsg = "already subscribed to a game"
}

func (s *session) handleMoveRequest(msgId types.ObjectId, req dataGamePlayerMoveRequest) {
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

	errMsg = "not allowed to move"
}

func (s *session) handleSendMsg(msgId types.ObjectId, req dataGameChatMsgSend) {
	var errMsg string
	defer func() {
		if errMsg != "" {
			s.sendErr(msgId, errMsg)
		}
	}()

	if s.sub != nil && s.gameId == req.GameID && s.userId == req.SenderID {
		if err := s.p.Publish(event.EventGameChatMsgeSent{
			ID:        req.ID,
			GameID:    req.GameID,
			SenderID:  req.SenderID,
			Content:   req.Content,
			Timestamp: time.Now().Unix(),
		}); err != nil {
			s.l.Error(fmt.Sprintf("failed to publish chat message event: %v", err))
		}
		return
	}

	errMsg = "not allowed to send message"
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
	return s.Conn.Close()
}

func (s *session) stop() {
	close(s.stopCh)
}

func (s *session) isClosed() bool {
	return s.closed.Load()
}

func (s *session) reconnect(conn *websocket.Conn) {
	s.closed.Store(false)
	s.stopCh = make(chan struct{})
	s.Conn = conn
}
