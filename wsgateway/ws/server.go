package ws

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"go.uber.org/atomic"

	"github.com/alikarimi999/shahboard/event"
	"github.com/alikarimi999/shahboard/pkg/log"
	"github.com/alikarimi999/shahboard/pkg/middleware"
	"github.com/alikarimi999/shahboard/types"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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
	sessions map[types.ObjectId][]*session
	counter  *atomic.Int64

	stopCh chan struct{}
	l      log.Logger
}

func NewServer(e *gin.Engine, s event.Subscriber, p event.Publisher,
	cfg *WsConfigs, l log.Logger) (*Server, error) {
	if cfg == nil {
		cfg = defaultConfigs
	}

	server := &Server{
		cfg:      cfg,
		sessions: make(map[types.ObjectId][]*session),
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

		sess := newSession(conn, user.ID, s, p, l)
		server.addSession(sess, user.ID)
		go server.sessionReader(sess)
		go server.sessionWriter(sess)
		sess.sendWelcome()

		l.Info(fmt.Sprintf("session '%d' created for user '%d'", sess.id, user.ID))

	})

	e.GET("/ws/reconnect/:id", middleware.ParsUserHeader(), func(ctx *gin.Context) {
		id := ctx.Param("id")
		sessId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "invalid session id"})
			return
		}

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

		sess, err := server.findSession(user.ID, types.ObjectId(sessId))
		if err != nil {
			fmt.Println(err)
			ctx.JSON(400, gin.H{"error": "session not found"})
			return
		}

		conn, er := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
		if er != nil {
			ctx.JSON(400, gin.H{"error": "failed to upgrade connection"})
			return
		}

		if !sess.reconnect(conn) {
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "session is open"))
			return
		}

		go server.sessionReader(sess)
		go server.sessionWriter(sess)
		sess.sendWelcome()

		l.Info(fmt.Sprintf("session '%d' reconnected for user '%d'", sess.id, user.ID))
	})

	return server, nil
}

func (s *Server) findSession(userId types.ObjectId, sessId types.ObjectId) (*session, error) {
	s.connsMux.RLock()
	defer s.connsMux.RUnlock()

	for _, sess := range s.sessions[userId] {
		if sess.id == sessId {
			return sess, nil
		}
	}

	return nil, fmt.Errorf("session not found")
}

func (s *Server) addSession(sess *session, userId types.ObjectId) {
	s.connsMux.Lock()
	defer s.connsMux.Unlock()

	s.sessions[userId] = append(s.sessions[userId], sess)
	s.counter.Add(1)
}

func (s *Server) removeSession(sess *session) {
	s.connsMux.Lock()
	defer s.connsMux.Unlock()

	for i, c := range s.sessions[sess.userId] {
		if c.id == sess.id {
			s.sessions[sess.userId] = append(s.sessions[sess.userId][:i], s.sessions[sess.userId][i+1:]...)
			s.counter.Add(-1)
			break
		}
	}
}

func (s *Server) stopSessions(remove bool, sess ...*session) {
	for _, se := range sess {
		if se.isClosed() {
			if remove {
				s.removeSession(se)
				s.l.Info(fmt.Sprintf("session '%d' removed from cache", se.id))
			}
			continue
		}
		if err := se.close(); err != nil {
			s.l.Error(fmt.Sprintf("session '%d' close error: '%s'", se.id, err))
			continue
		}
		s.l.Info(fmt.Sprintf("session '%d' closed", se.id))
		if remove {
			s.removeSession(se)
			s.l.Info(fmt.Sprintf("session '%d' removed from cache", se.id))
		}
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
		var msg ClientMsg
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

func (s *Server) isUserSubscribedToGame(user types.ObjectId) bool {
	s.connsMux.RLock()
	defer s.connsMux.RUnlock()

	for _, se := range s.sessions[user] {
		if se.isSubscribedToGame() {
			return true
		}
	}

	return false
}

func (s *Server) handleMsg(sess *session, msg *ClientMsg) {
	switch msg.Type {
	case MsgTypeFindMatch:
		if s.isUserSubscribedToGame(sess.userId) {
			sess.sendErr(msg.ID, "already subscribed to a game")
			return
		}

		d, ok := msg.Data.(DataFindMatchRequest)
		if !ok {
			sess.sendErr(msg.ID, "invalid data")
			return
		}

		sess.handlerFindMatchRequest(msg.ID, d)
	case MsgTypeView:
		d, ok := msg.Data.(DataGameViewRequest)
		if !ok {
			sess.sendErr(msg.ID, "invalid data")
			return
		}

		sess.handleViewGameRequest(msg.ID, d)
	case MsgTypeMove:
		d, ok := msg.Data.(DataGamePlayerMoveRequest)
		if !ok {
			sess.sendErr(msg.ID, "invalid data")
			return
		}

		sess.handleMoveRequest(msg.ID, d)

	case MsgTypePing:
		t := time.Now()
		sess.lastHeartBeat.Store(t)
		resp := ServerMsg{
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
