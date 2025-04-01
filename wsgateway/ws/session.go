package ws

import (
	"context"
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
	defaultViewGamesCap = 5
)

type session struct {
	*websocket.Conn
	id     types.ObjectId
	userId types.ObjectId

	eventCh chan event.Event
	msgCh   chan []byte
	pongCh  chan struct{}

	matchId types.ObjectId

	playGameId types.ObjectId

	vmu         sync.RWMutex
	viewGamesId map[types.ObjectId]struct{}
	viewCaps    int

	rc *redisCache

	h             *sessionsEventsHandler
	lastHeartBeat *atomic.Time

	stopCh chan struct{}

	p event.Publisher
	l log.Logger

	game GameService

	once    sync.Once
	stopped *atomic.Bool
}

func newSession(id types.ObjectId, conn *websocket.Conn, h *sessionsEventsHandler, rc *redisCache,
	userId types.ObjectId, gameId types.ObjectId, p event.Publisher, game GameService, l log.Logger) *session {
	s := &session{
		Conn:       conn,
		id:         id,
		userId:     userId,
		playGameId: gameId,

		eventCh: make(chan event.Event, 1000),
		msgCh:   make(chan []byte, 1000),
		pongCh:  make(chan struct{}),

		viewGamesId: make(map[types.ObjectId]struct{}),
		viewCaps:    defaultViewGamesCap,

		rc:            rc,
		h:             h,
		lastHeartBeat: atomic.NewTime(time.Now()),

		stopCh: make(chan struct{}),
		p:      p,
		l:      l,

		game:    game,
		stopped: atomic.NewBool(false),
	}
	go s.start()
	h.subscribeToBasicEvents(s)

	return s
}

func (s *session) start() {
	go s.handleEvent()
}

func (s *session) consume(e event.Event) {
	if s.isStopped() {
		return
	}

	select {
	case s.eventCh <- e:
	default:
		s.l.Error(fmt.Sprintf("session consume event failed: '%s'", e.GetTopic()))
	}
}

func (s *session) handleEvent() {
	for e := range s.eventCh {
		var msg *Msg
		switch e.GetTopic().Domain() {
		case event.DomainGame:
			msg = s.handleGameEvent(e)
		case event.DomainGameChat:
			msg = s.handleGameChatEvent(e)
		}

		if msg != nil {
			s.send(msg)
		}

	}
}

func (s *session) handleGameEvent(e event.Event) *Msg {
	var msg *Msg

	switch e.GetTopic().Action() {
	case event.ActionCreated:
		eve := e.(*event.EventGameCreated)

		s.playGameId = eve.GameID
		if err := s.rc.updateUserGameSession(context.Background(), s); err != nil {
			s.l.Error(err.Error())
			s.playGameId = types.ObjectZero

			return &Msg{
				MsgBase: MsgBase{
					Type:      MsgTypeError,
					Timestamp: time.Now().Unix(),
				},
				Data: []byte(MsgDataInternalErrorr),
			}
		}

		s.h.subscribeToGameWithChat(s, eve.GameID)
		msg = &Msg{
			MsgBase: MsgBase{
				Type:      MsgTypeGameCreate,
				Timestamp: time.Now().Unix(),
			},
			Data: eve.Encode(),
		}

	case event.ActionEnded:
		eve := e.(*event.EventGameEnded)
		msg = &Msg{
			MsgBase: MsgBase{
				Type:      MsgTypeGameEnd,
				Timestamp: time.Now().Unix(),
			},
			Data: e.Encode(),
		}
		if eve.GameID == s.playGameId {
			s.playGameId = types.ObjectZero
		}

		s.removeViewGames(eve.GameID)

		s.h.unsubscribeFromGameWithChat(s, eve.GameID)

	default:
		var mt MsgType
		switch e.GetTopic().Action() {
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
	switch e.GetTopic().Action() {
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

func (s *session) handleFindMatchRequest(msgId types.ObjectId, data dataFindMatchRequest) {
	var errMsg string
	defer func() {
		if errMsg != "" {
			s.sendErr(msgId, errMsg)
		}
	}()

	if s.playGameId.IsZero() && s.matchId.IsZero() && (s.userId == data.User1.ID || s.userId == data.User2.ID) {
		s.matchId = data.ID
		s.h.subscribeToMatch(s)
		return
	}

	errMsg = "already subscribed to a game"
}

func (s *session) handleResumeGameRequest(msgId types.ObjectId, req dataResumeGameRequest) {
	var errMsg string
	var msg *Msg
	defer func() {
		if errMsg != "" {
			s.sendErr(msgId, errMsg)
			return
		}
		if msg != nil {
			s.send(msg)
		}
	}()

	if s.matchId.IsZero() && s.playGameId.IsZero() {
		g, err := s.game.GetUserLiveGamePGN(context.Background(), s.userId)
		if err != nil {
			s.l.Error(err.Error())
			errMsg = MsgDataInternalErrorr
			return
		}

		if g == nil || g.GameId.String() != req.GameId.String() {
			errMsg = MsgDataNotFound
			return
		}

		s.playGameId = req.GameId
		if err := s.rc.updateUserGameSession(context.Background(), s); err != nil {
			s.l.Error(err.Error())
			s.playGameId = types.ObjectZero
			errMsg = MsgDataInternalErrorr
			return
		}

		s.h.subscribeToGameWithChat(s, req.GameId)
		if err := s.p.Publish(event.EventGamePlayerConnectionUpdated{
			GameID:    req.GameId,
			PlayerID:  s.userId,
			Connected: true,
			Timestamp: time.Now().Unix(),
		}); err != nil {
			s.l.Error(err.Error())
		}

		msg = &Msg{
			MsgBase: MsgBase{
				ID:        msgId,
				Type:      MsgTypeResumeGame,
				Timestamp: time.Now().Unix(),
			},
			Data: dataResumeGameResponse{
				GameId: req.GameId,
				Pgn:    g.Pgn,
			}.Encode(),
		}

		s.l.Debug(fmt.Sprintf("session '%s' subscribed to game '%s'", s.id, req.GameId))
		return
	}

	errMsg = "already subscribed to a game"
}

func (s *session) handleViewGameRequest(msgId types.ObjectId, req dataGameViewRequest) {
	var errMsg string
	var msg *Msg
	defer func() {
		if errMsg != "" {
			s.sendErr(msgId, errMsg)
			return
		}
		if msg != nil {
			s.send(msg)
		}
	}()

	s.addViewGame(req.GameId)
	game, err := s.game.GetLiveGamePGN(context.Background(), req.GameId)
	if err != nil {
		s.l.Error(err.Error())
		errMsg = MsgDataInternalErrorr
		s.removeViewGames(req.GameId)
		return
	}

	if game == nil || game.GameId.String() != req.GameId.String() {
		errMsg = MsgDataNotFound
		s.removeViewGames(req.GameId)
		return
	}

	msg = &Msg{
		MsgBase: MsgBase{
			ID:        msgId,
			Type:      MsgTypeViewGame,
			Timestamp: time.Now().Unix(),
		},
		Data: dataGameViewResponse{
			GameId: req.GameId,
			Pgn:    game.Pgn,
		}.Encode(),
	}

	s.h.subscribeToGameWithChat(s, req.GameId)

	if err := s.rc.addToGameViwersList(context.Background(), s.userId, req.GameId); err != nil {
		s.l.Error(err.Error())
	}

	s.l.Debug(fmt.Sprintf("session '%s' subscribed to game '%s' as viewer", s.id, req.GameId))
}

func (s *session) handleMoveRequest(msgId types.ObjectId, req dataGamePlayerMoveRequest) {
	var errMsg string
	defer func() {
		if errMsg != "" {
			s.sendErr(msgId, errMsg)
		}
	}()

	if s.userId == req.PlayerID && s.playGameId == req.GameID {
		if err := s.p.Publish(req.EventGamePlayerMoved); err != nil {
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

	if s.userId == req.SenderID && s.playGameId == req.GameID {
		if err := s.p.Publish(req.EventGameChatMsgeSent); err != nil {
			s.l.Error(fmt.Sprintf("failed to publish chat message event: %v", err))
		}
		return
	}

	errMsg = "not allowed to send message"
}

func (s *session) send(msg *Msg) {
	select {
	case s.msgCh <- msg.Encode():
	default:
		s.l.Error("failed to send message: channel is full")
	}
}

func (s *session) sendErr(id types.ObjectId, err string) {
	s.send(&Msg{
		MsgBase: MsgBase{
			ID:        id,
			Type:      MsgTypeError,
			Timestamp: time.Now().Unix(),
		},
		Data: []byte(err),
	})
}

func (s *session) sendWelcome() {
	s.send(&Msg{
		MsgBase: MsgBase{
			ID:        types.NewObjectId(),
			Type:      MsgTypeWelcome,
			Timestamp: time.Now().Unix(),
		},
		Data: []byte("welcome"),
	})
}

func (s *session) sendViwersList(gameId types.ObjectId, viewers []types.ObjectId) {
	s.send(&Msg{
		MsgBase: MsgBase{
			ID:        types.NewObjectId(),
			Type:      MsgTypeViewersList,
			Timestamp: time.Now().Unix(),
		},
		Data: dataViwersListResponse{
			GameId: gameId,
			List:   viewers,
		}.Encode(),
	})
}

func (s *session) sendPong() {
	s.pongCh <- struct{}{}
}

func (s *session) isStopped() bool {
	return s.stopped.Load()
}

func (s *session) stop() {
	s.once.Do(func() {
		s.stopped.Store(true)
		close(s.eventCh)
		close(s.msgCh)
		close(s.stopCh)

		var gamesId []types.ObjectId
		if !s.playGameId.IsZero() {
			gamesId = append(gamesId, s.playGameId)
		}

		ids := s.removeAllViewGames()
		if len(ids) > 0 {
			if err := s.rc.removeFromGameViewersList(context.Background(), s.userId, ids...); err != nil {
				s.l.Error(err.Error())
			}
			gamesId = append(gamesId, ids...)
		}

		s.h.unsubscribeFromBasicEvents(s)
		s.h.unsubscribeFromGameWithChat(s, gamesId...)

		if !s.matchId.IsZero() {
			s.h.unsubscribeFromMatch(s)
		}
	})
}

func (s *session) addViewGame(gameId types.ObjectId) (msg string) {
	s.vmu.Lock()
	defer s.vmu.Unlock()

	if _, ok := s.viewGamesId[gameId]; ok {
		return "already subscribed to this game"
	}

	if len(s.viewGamesId) >= s.viewCaps {
		return "view cap reached"
	}

	s.viewGamesId[gameId] = struct{}{}

	return ""
}

func (s *session) getAllViewGames() []types.ObjectId {
	s.vmu.RLock()
	defer s.vmu.RUnlock()

	var gamesId []types.ObjectId
	for id := range s.viewGamesId {
		gamesId = append(gamesId, id)
	}

	return gamesId
}

func (s *session) removeViewGames(gameId types.ObjectId) {
	s.vmu.Lock()
	defer s.vmu.Unlock()

	delete(s.viewGamesId, gameId)
}

func (s *session) removeAllViewGames() []types.ObjectId {
	s.vmu.Lock()
	defer s.vmu.Unlock()

	var gamesId []types.ObjectId
	for id := range s.viewGamesId {
		gamesId = append(gamesId, id)
	}

	s.viewGamesId = make(map[types.ObjectId]struct{})
	return gamesId
}
