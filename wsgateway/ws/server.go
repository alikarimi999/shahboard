package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/alikarimi999/shahboard/event"
	"github.com/alikarimi999/shahboard/pkg/jwt"
	"github.com/alikarimi999/shahboard/pkg/log"
	"github.com/alikarimi999/shahboard/pkg/middleware"
	"github.com/alikarimi999/shahboard/types"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Server is an implementation of entity.StreamIn
type Server struct {
	cfg *WsConfigs

	sm *sessionsManager
	em *endedGamesList

	h *sessionsEventsHandler

	p     event.Publisher
	cache *redisCache

	stopCh chan struct{}

	jwtValidator *jwt.Validator

	l log.Logger
}

func NewServer(e *gin.Engine, s event.Subscriber, p event.Publisher, game GameService,
	cfg *WsConfigs, c *redis.Client, v *jwt.Validator, l log.Logger) (*Server, error) {
	if cfg == nil {
		cfg = defaultConfigs
	}

	em := newEndedGamesList()
	server := &Server{
		cfg:          cfg,
		cache:        newRedisCache(c, cfg.UserSessionsCap, l),
		sm:           newSessionsManager(),
		em:           em,
		h:            newSessionsEventsHandler(s, em, l),
		p:            p,
		stopCh:       make(chan struct{}),
		jwtValidator: v,
		l:            l,
	}

	go server.manageSessionsState()

	e.GET("/ws", middleware.ParseQueryToken(v), func(ctx *gin.Context) {
		user, ok := middleware.ExtractUser(ctx)
		if !ok {
			ctx.JSON(400, gin.H{"error": "failed to parse token"})
			ctx.Abort()
			return
		}

		conn, er := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
		if er != nil {
			ctx.JSON(400, gin.H{"error": "failed to upgrade connection"})
			ctx.Abort()
			return
		}

		id := types.NewObjectId()
		success, err := server.cache.addUserSessionId(context.Background(), user.ID, id)
		if err != nil {
			l.Error(fmt.Sprintf("failed to save session state: %v", err))
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "internal server error"))
			conn.Close()
			return
		}

		if !success {
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "open sessions cap reached"))
			conn.Close()
			return
		}

		sess := newSession(id, conn, server.h, server.cache, user.ID, types.ObjectZero, p, game, l)
		server.sm.add(sess)

		go server.sessionReader(sess)
		go server.sessionWriter(sess)
		sess.sendWelcome()

		l.Debug(fmt.Sprintf("session '%s' created for '%s'", sess.id, user.ID))

	})

	return server, nil
}

func (s *Server) stopSessions(ss ...*session) {
	s.sm.remove(ss...)
	events := []event.Event{}
	usersGameId := make(map[types.ObjectId]types.ObjectId)

	for _, se := range ss {
		se.stop()

		if !se.playGameId.Load().IsZero() {
			counter, err := s.cache.removeGameIdFromUserSessions(context.Background(), se.userId, se.id)
			if err != nil {
				// TODO: need to handle this error
				s.l.Error(err.Error())

			} else if counter == 0 {
				// this is the last session for this user that is playing the game
				// so we need to notify other parts of system
				usersGameId[se.userId] = se.playGameId.Load()
			}
		}
	}

	if err := s.cache.deleteUsersSessions(context.Background(), ss...); err != nil {
		s.l.Error(err.Error())
	}

	for u, g := range usersGameId {
		events = append(events, event.EventGamePlayerLeft{
			GameID:    g,
			PlayerID:  u,
			Timestamp: time.Now().Unix(),
		})
	}
	if len(events) > 0 {
		if err := s.p.Publish(events...); err != nil {
			s.l.Error(err.Error())
		}
	}
}

func (s *Server) sessionReader(se *session) {
	// defer func() {
	// 	s.l.Debug(fmt.Sprintf("session '%s' reader stopped", se.id))
	// }()

	for {
		if se.isStopped() {
			// this sholudn't happen
			s.l.Debug(fmt.Sprintf("read from stopped session: %s", se.id))
			return
		}

		mt, recievedMsg, err := se.ReadMessage()
		if err != nil {
			if !se.isStopped() {
				// s.l.Debug(fmt.Sprintf("session '%s' read message error: %v", se.id, err))
				s.stopSessions(se)
			}
			return
		}

		if mt == websocket.BinaryMessage && len(recievedMsg) > 0 && recievedMsg[0] == 0x0 {
			se.lastHeartBeat.Store(time.Now())
			se.sendPong()
			continue
		}

		if mt == websocket.TextMessage && len(recievedMsg) > 0 {
			var msg Msg
			if err := json.Unmarshal(recievedMsg, &msg); err != nil {
				continue
			}
			s.handleMsg(se, &msg)
		}
	}
}

func (s *Server) sessionWriter(se *session) {
	// defer func() {
	// 	s.l.Debug(fmt.Sprintf("session '%s' writer stopped", se.id))
	// }()

	for {
		select {
		case <-se.stopCh:
			return
		case message := <-se.msgCh:
			if message == nil || se.isStopped() {
				return
			}

			if err := se.WriteMessage(websocket.TextMessage, message); err != nil {
				if !se.isStopped() {
					// s.l.Debug(fmt.Sprintf("session '%s' write message error: %v", se.id, err))
					s.stopSessions(se)
				}
				return
			}
		case <-se.pongCh:
			if err := se.WriteMessage(websocket.BinaryMessage, []byte{0x1}); err != nil {
				if !se.isStopped() {
					s.l.Debug(fmt.Sprintf("session '%s' write pong message error: %v", se.id, err))
					s.stopSessions(se)
				}
				return
			}
		}
	}
}

func (s *Server) handleMsg(sess *session, msg *Msg) {
	switch msg.Type {
	case MsgTypeFindMatch:
		var d DataFindMatchRequest
		if err := json.Unmarshal(msg.Data, &d); err != nil {
			sess.sendErr(msg.ID, "invalid data")
			return
		}

		sess.handleFindMatchRequest(msg.ID, d)
	case MsgTypeResumeGame:
		var d DataResumeGameRequest
		if err := json.Unmarshal(msg.Data, &d); err != nil {
			sess.sendErr(msg.ID, "invalid data")
			return
		}

		sess.handleResumeGameRequest(msg.ID, d)
	case MsgTypeViewGame:
		var d DataGameViewRequest
		if err := json.Unmarshal(msg.Data, &d); err != nil {
			sess.sendErr(msg.ID, "invalid data")
			return
		}

		sess.handleViewGameRequest(msg.ID, d)
	case MsgTypePlayerMove:
		var d DataGamePlayerMoveRequest
		if err := json.Unmarshal(msg.Data, &d); err != nil {
			sess.sendErr(msg.ID, "invalid data")
			return
		}

		sess.handleMoveRequest(msg.ID, d)
	case MsgTypePlayerResigned:
		var d DataGamePlayerResignRequest
		if err := json.Unmarshal(msg.Data, &d); err != nil {
			sess.sendErr(msg.ID, "invalid data")
			return
		}

		sess.handlePlayerResignRequest(msg.ID, d)
	case MsgTypeChatMsgSend:
		var d DataGameChatMsgSend
		if err := json.Unmarshal(msg.Data, &d); err != nil {
			sess.sendErr(msg.ID, "invalid data")
			return
		}

		sess.handleSendMsg(msg.ID, d)
	case MsgTypeData:

		// handle data message
	}
}

func (s *Server) GetLiveGamesViewersNumber(ctx context.Context) (map[types.ObjectId]int64, error) {
	games, err := s.cache.countAllGamesViewers(ctx)
	if err != nil {
		return nil, err
	}

	return games, nil
}
