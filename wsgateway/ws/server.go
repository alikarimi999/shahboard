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
		sessions: make(map[types.ObjectId]*session),
		counter:  atomic.NewInt64(0),
		stopCh:   make(chan struct{}),
		l:        l,
	}

	go server.checkHeartbeat()

	e.GET("/ws", middleware.ParsUserHeader(), func(ctx *gin.Context) {
		u, _ := ctx.Get("user")
		user, ok := u.(types.User)
		if !ok {
			ctx.JSON(401, gin.H{"error": "unauthorized"})
			return
		}

		if user.ID == 0 {
			ctx.JSON(401, gin.H{"error": "unauthorized"})
			return
		}

		conn, er := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
		if er != nil {
			ctx.JSON(400, gin.H{"error": "failed to upgrade connection"})
			return
		}

		// check if user has a session in memory
		sess := server.getSession(user.ID)
		if sess != nil {
			if sess.isClosed() {
				sess.reconnect(conn)
				if err := server.cache.SaveSessionState(context.Background(), sess); err != nil {
					l.Error(fmt.Sprintf("failed to save session state: %v", err))
					conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "internal server error"))
					conn.Close()
					return
				}
				go server.sessionReader(sess)
				go server.sessionWriter(sess)
				sess.sendWelcome()
				l.Info(fmt.Sprintf("session '%d' reconnected for user '%d'", sess.id, user.ID))

				return
			}
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "session is open"))
			conn.Close()
			return
		}

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
		// if csess != nil && csess.Closed {
		// 	server.cache.DeleteSession(context.Background(), user.ID)
		// }

		sess = newSession(conn, newGameEventsManager(s, l), user.ID, p, l)
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

		l.Info(fmt.Sprintf("session '%d' created for user '%d'", sess.id, user.ID))

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

func (s *Server) stopSessions(remove bool, sess ...*session) {
	for _, se := range sess {
		if se.isClosed() {
			if remove {
				s.removeSession(se)
			}
			continue
		}
		if err := se.close(); err != nil {
			s.l.Error(fmt.Sprintf("session '%d' close error: '%s'", se.id, err))
			continue
		}
		s.l.Debug(fmt.Sprintf("session '%d' closed", se.id))
		if remove {
			s.removeSession(se)
			s.l.Debug(fmt.Sprintf("session '%d' removed from cache", se.id))
		}
	}

	if len(sess) > 0 {
		// Preserves session state in Redis by marking as 'closed' rather than deleting
		// - Enables session resumption across WS gateway instances by retaining full state data
		// - Maintains cluster-wide visibility of disconnected sessions for coordination
		// - Supports future session recovery workflows without data reconstruction
		if err := s.cache.SaveSessionsState(context.Background(), sess...); err != nil {
			s.l.Error(fmt.Sprintf("failed to save sessions state: %v", err))
		}
		ids := []types.ObjectId{}
		for _, s := range sess {
			ids = append(ids, s.userId)
		}
		s.l.Debug(fmt.Sprintf("sessions '%v' closed on redis", ids))
	}
}

func (s *Server) sessionReader(sess *session) {
	defer func() {
		s.l.Debug(fmt.Sprintf("session '%d' reader stopped", sess.id))
	}()

	for {
		_, recievedMsg, err := sess.ReadMessage()
		if err != nil {
			s.stopSessions(false, sess)
			return
		}
		var msg Msg
		if err := json.Unmarshal(recievedMsg, &msg); err != nil {
			continue
		}

		s.handleMsg(sess, &msg)
	}
}

func (s *Server) sessionWriter(sess *session) {
	defer func() {
		s.l.Debug(fmt.Sprintf("session '%d' writer stopped", sess.id))
	}()

	for {
		select {
		case <-sess.stopCh:
			return
		case message := <-sess.msgCh:
			if err := sess.WriteMessage(websocket.TextMessage, message); err != nil {
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

	case MsgTypePing:
		t := time.Now()
		sess.lastHeartBeat.Store(t)
		resp := Msg{
			MsgBase: MsgBase{
				ID:        msg.ID,
				Timestamp: t.Unix(),
				Type:      MsgTypePong,
			}}
		sess.send(resp)
	case MsgTypeData:

		// handle data message
	}
}
