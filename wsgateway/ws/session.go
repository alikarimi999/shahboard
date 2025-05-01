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

// TODO: implement session in a separate package for better modularity
type session struct {
	*websocket.Conn
	id      types.ObjectId
	userId  types.ObjectId
	isGuest bool

	eventCh chan event.Event
	msgCh   chan []byte
	pongCh  chan struct{}

	// These variables may be updated multiple times during the session's lifetime.
	// Access must be concurrency-safe to prevent race conditions.
	matchId    *types.AtomicObjectId
	playGameId *types.AtomicObjectId

	vmu         sync.RWMutex
	viewGamesId map[types.ObjectId]struct{}
	viewCaps    int

	rc *redisCache

	h             *sessionsEventsHandler
	lastHeartBeat *atomic.Time

	p event.Publisher
	l log.Logger

	game GameService

	cleanUP func(*session)
	once    sync.Once
	wg      sync.WaitGroup
	stopCh  chan struct{}
}

func newSession(id types.ObjectId, conn *websocket.Conn, h *sessionsEventsHandler, rc *redisCache,
	userId types.ObjectId, isGuest bool, gameId types.ObjectId, p event.Publisher,
	game GameService, l log.Logger, cleanUP func(*session)) *session {
	s := &session{
		Conn:    conn,
		id:      id,
		userId:  userId,
		isGuest: isGuest,

		matchId:    types.NewAtomicObjectId(types.ObjectZero),
		playGameId: types.NewAtomicObjectId(gameId),

		eventCh: make(chan event.Event, 10),
		msgCh:   make(chan []byte, 10),
		pongCh:  make(chan struct{}),

		viewGamesId: make(map[types.ObjectId]struct{}),
		viewCaps:    defaultViewGamesCap,

		rc:            rc,
		h:             h,
		lastHeartBeat: atomic.NewTime(time.Now()),

		p: p,
		l: l,

		game: game,

		stopCh:  make(chan struct{}),
		cleanUP: cleanUP,
	}
	go s.start()
	h.subscribeToBasicEvents(s)

	return s
}

func (s *session) start() {
	go s.readLoop()
	go s.writeLoop()
	go s.handleEvent()
	s.sendWelcome()
}

func (s *session) consume(e event.Event) {
	defer func() {
		if r := recover(); r != nil {
			s.l.Warn(fmt.Sprintf("session '%s' consume(e event.Event) panicked: %v", s.id, r))
		}
	}()

	s.eventCh <- e
}

func (s *session) handleEvent() {
	s.wg.Add(1)
	defer s.wg.Done()
	for {
		select {
		case <-s.stopCh:
			return
		case e := <-s.eventCh:
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
}

func (s *session) handleGameEvent(e event.Event) *Msg {
	var msg *Msg

	switch e.GetTopic().Action() {
	case event.ActionCreated:
		eve := e.(*event.EventGameCreated)

		s.playGameId.Store(eve.GameID)
		if err := s.rc.addGameIdToUserSessions(context.Background(), s.userId, s.id, eve.GameID); err != nil {
			s.l.Error(err.Error())
			s.playGameId.SetZero()

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
		if eve.GameID == s.playGameId.Load() {
			s.playGameId.SetZero()
		}

		if _, err := s.rc.removeGameIdFromUserSessions(context.Background(), s.userId, s.id); err != nil {
			// TODO: handle error
			s.l.Error(err.Error())
		}
		s.removeViewGames(eve.GameID)
		s.h.unsubscribeFromGameWithChat(s, eve.GameID)

	default:
		var mt MsgType
		switch e.GetTopic().Action() {
		case event.ActionGameMoveApprove:
			mt = MsgTypeMoveApproved
		case event.ActionGamePlayerJoined:
			mt = MsgTypePlayerJoined
		case event.ActionGamePlayerLeft:
			mt = MsgTypePlayerLeft
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

func (s *session) handleFindMatchRequest(msgId types.ObjectId, data DataFindMatchRequest) {
	var errMsg string
	defer func() {
		if errMsg != "" {
			s.sendErr(msgId, errMsg)
		}
	}()

	// if s.isGuest {
	// 	errMsg = guestNotAllowed
	// 	return
	// }

	if s.playGameId.Load().IsZero() && s.matchId.Load().IsZero() &&
		(s.userId == data.User1.ID || s.userId == data.User2.ID) {
		s.matchId.Store(data.ID)
		s.h.subscribeToMatch(s)
		return
	}

	errMsg = "already subscribed to a game"
}

func (s *session) handleResumeGameRequest(msgId types.ObjectId, req DataResumeGameRequest) {
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

	// if s.isGuest {
	// 	errMsg = guestNotAllowed
	// 	return
	// }

	if s.matchId.Load().IsZero() && s.playGameId.Load().IsZero() {
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

		s.playGameId.Store(req.GameId)
		if err := s.rc.addGameIdToUserSessions(context.Background(), s.userId, s.id, req.GameId); err != nil {
			s.l.Error(err.Error())
			s.playGameId.SetZero()
			errMsg = MsgDataInternalErrorr
			return
		}

		s.h.subscribeToGameWithChat(s, req.GameId)
		if err := s.p.Publish(event.EventGamePlayerJoined{
			GameID:    req.GameId,
			PlayerID:  s.userId,
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
			Data: DataResumeGameResponse{
				GameId:               req.GameId,
				Pgn:                  g.Pgn,
				PlayersDisconnection: g.PlayersDisconnections,
			}.Encode(),
		}

		// s.l.Debug(fmt.Sprintf("session '%s' subscribed to game '%s'", s.id, req.GameId))
		return
	}

	errMsg = "already subscribed to a game"
}

func (s *session) handleViewGameRequest(msgId types.ObjectId, req DataGameViewRequest) {
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

	emsg := s.addViewGame(req.GameId)
	if emsg != "" {
		errMsg = emsg
		return
	}
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
		Data: DataGameViewResponse{
			GameId:               req.GameId,
			Pgn:                  game.Pgn,
			PlayersDisconnection: game.PlayersDisconnections,
		}.Encode(),
	}

	s.h.subscribeToGameWithChat(s, req.GameId)

	// Offload adding game viewer lists to the cache to a server-side worker using batching,
	// instead of handling it here for every individual connection.

	// if err := s.rc.addToGameViwersList(context.Background(), s.userId, req.GameId); err != nil {
	// 	s.l.Error(err.Error())
	// }

	// s.l.Debug(fmt.Sprintf("session '%s' subscribed to game '%s' as viewer", s.id, req.GameId))
}

func (s *session) handleMoveRequest(msgId types.ObjectId, req DataGamePlayerMoveRequest) {
	var errMsg string
	defer func() {
		if errMsg != "" {
			s.sendErr(msgId, errMsg)
		}
	}()

	if s.userId == req.PlayerID && s.playGameId.Load() == req.GameID {
		if err := s.p.Publish(req.EventGamePlayerMoved); err != nil {
			s.l.Error(fmt.Sprintf("failed to publish move event: %v", err))
		}
		return
	}

	errMsg = "not allowed to move"
}

func (s *session) handlePlayerResignRequest(msgId types.ObjectId, req DataGamePlayerResignRequest) {
	var errMsg string
	defer func() {
		if errMsg != "" {
			s.sendErr(msgId, errMsg)
		}
	}()

	if s.userId == req.PlayerID && s.playGameId.Load() == req.GameID {
		if err := s.p.Publish(req.EventGamePlayerResigned); err != nil {
			s.l.Error(fmt.Sprintf("failed to publish resign event: %v", err))
		}
		return
	}

	errMsg = "not allowed to resign"
}

func (s *session) handleSendMsg(msgId types.ObjectId, req DataGameChatMsgSend) {
	var errMsg string
	defer func() {
		if errMsg != "" {
			s.sendErr(msgId, errMsg)
		}
	}()

	if s.userId == req.SenderID && s.playGameId.Load() == req.GameID {
		if err := s.p.Publish(req.EventGameChatMsgeSent); err != nil {
			s.l.Error(fmt.Sprintf("failed to publish chat message event: %v", err))
		}
		return
	}

	errMsg = "not allowed to send message"
}

func (s *session) send(msg *Msg) {
	defer func() {
		if r := recover(); r != nil {
			s.l.Warn(fmt.Sprintf("session '%s' send(msg *Msg) panicked: %v", s.id, r))
		}
	}()

	s.msgCh <- msg.Encode()
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

// func (s *session) sendViwersList(gameId types.ObjectId, viewers []types.ObjectId) {
// 	s.send(&Msg{
// 		MsgBase: MsgBase{
// 			ID:        types.NewObjectId(),
// 			Type:      MsgTypeViewersList,
// 			Timestamp: time.Now().Unix(),
// 		},
// 		Data: DataViwersListResponse{
// 			GameId: gameId,
// 			List:   viewers,
// 		}.Encode(),
// 	})
// }

func (s *session) sendPong() {
	// not sure recover is needed (just for test and expriment!)
	defer func() {
		if r := recover(); r != nil {
			s.l.Warn(fmt.Sprintf("session '%s' sendPong() panicked: %v", s.id, r))
		}
	}()

	// to make sure it won't cause deadlock in stopping session which is a rare deadlock
	select {
	case s.pongCh <- struct{}{}:
	default:
	}
}

func (s *session) Stop() {
	s.once.Do(func() {
		go func() {
			s.h.unsubscribeFromBasicEvents(s)
			if !s.matchId.Load().IsZero() {
				s.h.unsubscribeFromMatch(s)
			}

			var gamesId []types.ObjectId
			if !s.playGameId.Load().IsZero() {
				gamesId = append(gamesId, s.playGameId.Load())
			}

			gamesId = append(gamesId, s.getAllViewGames()...)
			s.h.unsubscribeFromGameWithChat(s, gamesId...)

			if err := s.Conn.Close(); err != nil {
				s.l.Error(err.Error())
			}

			close(s.stopCh)

			s.wg.Wait()
			close(s.eventCh)
			close(s.msgCh)
			close(s.pongCh)

			s.l.Debug(fmt.Sprintf("session '%s' stopped for user '%s'", s.id, s.userId))

			s.cleanUP(s)
		}()
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
