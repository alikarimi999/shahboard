package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/atomic"

	"github.com/alikarimi999/shahboard/event"
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

	connsMux sync.RWMutex
	sessions map[types.ObjectId]*session
	counter  *atomic.Int64

	m *gameEventsManager

	p event.Publisher

	cache *redisCache

	stopCh chan struct{}
	l      log.Logger
}

func NewServer(e *gin.Engine, s event.Subscriber, p event.Publisher,
	cfg *WsConfigs, c *redis.Client, l log.Logger) (*Server, error) {
	if cfg == nil {
		cfg = defaultConfigs
	}

	server := &Server{
		cfg:      cfg,
		cache:    newRedisCache(c),
		m:        newGameEventsManager(s, l),
		p:        p,
		sessions: make(map[types.ObjectId]*session),
		counter:  atomic.NewInt64(0),
		stopCh:   make(chan struct{}),
		l:        l,
	}

	go server.checkHeartbeat()

	e.GET("/ws", middleware.ParsQueryToken(), func(ctx *gin.Context) {
		u, _ := ctx.Get("user")
		user := u.(types.User)

		conn, er := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
		if er != nil {
			ctx.JSON(400, gin.H{"error": "failed to upgrade connection"})
			ctx.Abort()
			return
		}

		// // check if user has a session in memory
		// sess := server.getSession(user.ID)
		// if sess != nil {
		// 	if sess.isClosed() {
		// 		sess.reconnect(conn)
		// 		if err := server.cache.SaveSessionState(context.Background(), sess); err != nil {
		// 			l.Error(fmt.Sprintf("failed to save session state: %v", err))
		// 			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "internal server error"))
		// 			conn.Close()
		// 			return
		// 		}
		// 		go server.sessionReader(sess)
		// 		go server.sessionWriter(sess)
		// 		sess.sendWelcome()
		// 		l.Info(fmt.Sprintf("session '%s' reconnected for user '%s'", sess.id, user.ID))

		// 		return
		// 	}
		// 	conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "session is open"))
		// 	conn.Close()
		// 	return
		// }

		// check if user has a session in redis
		csess, err := server.cache.GetSession(context.Background(), user.ID)
		if err != nil {
			l.Error(fmt.Sprintf("failed to check if user has session: %v", err))
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "internal server error"))
			conn.Close()
			return
		}

		// this means that user has an open session in another wsgateway instance
		if csess != nil && !csess.Closed {
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "session is open"))
			conn.Close()
			return
		}

		// this means that user has a closed session in another wsgateway instance
		// don't need to save session state if it's closed,it will be replaced by a new session
		// in future we should be able to retrieve the session state from redis and reconnect the session instead of creating a new one
		if csess != nil && csess.Closed {
			msgs, err := server.cache.GetSessionMsgs(context.Background(), csess.SessionId)
			if err != nil {
				l.Error(fmt.Sprintf("failed to get session messages: %v", err))
				conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "internal server error"))
				conn.Close()
				return
			}

			server.l.Debug(fmt.Sprintf("delete disconnected session for user '%s'", user.ID))
			sess := server.getSession(user.ID)
			if sess != nil {
				server.stopSessions(true, sess)
			}

			sess = newSession(conn, newGameEventsManager(s, l), server.cache, user.ID, p, l)
			if err := server.cache.SaveSessionState(context.Background(), sess); err != nil {
				l.Error(fmt.Sprintf("failed to save session state: %v", err))
				conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "internal server error"))
				conn.Close()
				return
			}

			server.addSession(sess, user.ID)
			go server.sessionReader(sess)
			go server.sessionWriter(sess)
			sess.sendWelcome()

			for _, msg := range msgs {
				sess.send(msg)
			}

			if err := sess.p.Publish(event.EventGamePlayerConnectionUpdated{
				ID:        types.NewObjectId(),
				GameID:    csess.GameId,
				PlayerID:  user.ID,
				Connected: true,
				Timestamp: time.Now().Unix(),
			}); err != nil {
				l.Error(fmt.Sprintf("failed to publish game_player_connection_updated event: %v", err))
			}

			if !csess.GameId.IsZero() {
				if csess.Role == gamePlayerRole {
					sess.SubscribeAsPlayer(csess.GameId)
				} else if csess.Role == gameViewerRole {
					sess.SubscribeAsViewer(csess.GameId)
				}
			}

			l.Info(fmt.Sprintf("new session '%s' created for user '%s'", sess.id, user.ID))
			return
		}

		sess := newSession(conn, newGameEventsManager(s, l), server.cache, user.ID, p, l)
		if err := server.cache.SaveSessionState(context.Background(), sess); err != nil {
			l.Error(fmt.Sprintf("failed to save session state: %v", err))
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "internal server error"))
			conn.Close()
			return
		}

		server.addSession(sess, user.ID)

		go server.sessionReader(sess)
		go server.sessionWriter(sess)
		sess.sendWelcome()

		l.Info(fmt.Sprintf("session '%s' created for user '%s'", sess.id, user.ID))

	})

	return server, nil
}

func (s *Server) getSession(userId types.ObjectId) *session {
	s.connsMux.RLock()
	defer s.connsMux.RUnlock()

	return s.sessions[userId]
}

func (s *Server) addSession(sess *session, userId types.ObjectId) {
	s.connsMux.Lock()
	defer s.connsMux.Unlock()

	s.sessions[userId] = sess
	s.counter.Add(1)
}

func (s *Server) removeSession(sess *session) {
	s.connsMux.Lock()
	defer s.connsMux.Unlock()

	delete(s.sessions, sess.userId)
	s.counter.Add(-1)
}

func (s *Server) stopSessions(purge bool, sess ...*session) {
	for _, se := range sess {
		if se.isClosed() {
			if purge {
				s.removeSession(se)
				se.stop()
			}
			continue
		}
		if err := se.close(); err != nil {
			s.l.Error(fmt.Sprintf("session '%s' close error: '%s'", se.id, err))
			continue
		}
		s.l.Debug(fmt.Sprintf("session '%s' closed", se.id))
		if purge {
			s.removeSession(se)
			se.stop()
			s.l.Debug(fmt.Sprintf("session '%s' removed from cache", se.id))
		}
	}

	if len(sess) == 0 {
		return
	}

	if purge {
		ids := []types.ObjectId{}
		for _, s := range sess {
			ids = append(ids, s.userId)
		}
		if err := s.cache.DeleteSessions(context.Background(), ids...); err != nil {
			s.l.Error(fmt.Sprintf("failed to delete sessions: %v", err))
			return
		}
		s.l.Debug(fmt.Sprintf("user's sessions closed on redis: '%v'", ids))
		return
	}

	events := []event.Event{}
	t := time.Now().Unix()
	for _, se := range sess {
		if !se.gameId.IsZero() {
			events = append(events, event.EventGamePlayerConnectionUpdated{
				ID:        types.NewObjectId(),
				GameID:    se.gameId,
				PlayerID:  se.userId,
				Connected: false,
				Timestamp: t,
			})
		}
	}

	if len(events) > 0 {
		if err := s.p.Publish(events...); err != nil {
			s.l.Error(fmt.Sprintf("failed to publish game_player_connection_updated event: %v", err))
		}
	}

	if err := s.cache.SaveSessionsState(context.Background(), sess...); err != nil {
		s.l.Error(fmt.Sprintf("failed to save sessions state: %v", err))
		return
	}
	ids := []types.ObjectId{}
	for _, s := range sess {
		ids = append(ids, s.userId)
	}
	s.l.Debug(fmt.Sprintf("sessions '%v' closed on redis", ids))
}

func (s *Server) sessionReader(sess *session) {
	defer func() {
		s.l.Debug(fmt.Sprintf("session '%s' reader stopped", sess.id))
	}()

	for {
		mt, recievedMsg, err := sess.ReadMessage()
		if err != nil {
			s.l.Debug(fmt.Sprintf("session '%s' read message error: %v", sess.id, err))
			// s.stopSessions(false, sess)
			return
		}

		if mt == websocket.BinaryMessage && len(recievedMsg) > 0 && recievedMsg[0] == 0x0 {
			sess.lastHeartBeat.Store(time.Now())
			sess.sendPong()
			continue
		}

		if mt == websocket.TextMessage && len(recievedMsg) > 0 {
			var msg Msg
			if err := json.Unmarshal(recievedMsg, &msg); err != nil {
				continue
			}
			s.handleMsg(sess, &msg)
		}
	}
}

func (s *Server) sessionWriter(sess *session) {
	defer func() {
		s.l.Debug(fmt.Sprintf("session '%s' writer stopped", sess.id))
	}()

	for {
		select {
		case <-sess.stopCh:
			return
		case message := <-sess.msgCh:
			if err := sess.WriteMessage(websocket.TextMessage, message); err != nil {
				s.l.Debug(fmt.Sprintf("session '%s' write message error: %v", sess.id, err))
				s.stopSessions(false, sess)
				return
			}
		case <-sess.pongCh:
			if err := sess.WriteMessage(websocket.BinaryMessage, []byte{0x1}); err != nil {
				s.l.Debug(fmt.Sprintf("session '%s' write pong message error: %v", sess.id, err))
				s.stopSessions(false, sess)
				return
			}
		}
	}
}

func (s *Server) handleMsg(sess *session, msg *Msg) {
	switch msg.Type {
	case MsgTypeFindMatch:
		if sess.isSubscribedToGame() {
			sess.sendErr(msg.ID, "already subscribed to a game")
			return
		}

		var d DataFindMatchRequest
		if err := json.Unmarshal(msg.Data, &d); err != nil {
			sess.sendErr(msg.ID, "invalid data")
			return
		}

		sess.handleFindMatchRequest(msg.ID, d)
	case MsgTypeView:
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

	case MsgTypeData:

		// handle data message
	}
}
