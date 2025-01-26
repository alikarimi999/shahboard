package ws

import (
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

		sess := server.findSession(user.ID)
		if sess != nil {
			if sess.isClosed() {
				sess.reconnect(conn)
				go server.sessionReader(sess)
				go server.sessionWriter(sess)
				sess.sendWelcome()
				l.Info(fmt.Sprintf("session '%d' reconnected for user '%d'", sess.id, user.ID))

				return
			}
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "session is open"))
			return
		}

		sess = newSession(conn, user.ID, s, p, l)
		server.addSession(sess, user.ID)
		go server.sessionReader(sess)
		go server.sessionWriter(sess)
		sess.sendWelcome()

		l.Info(fmt.Sprintf("session '%d' created for user '%d'", sess.id, user.ID))

	})

	return server, nil
}

func (s *Server) findSession(userId types.ObjectId) *session {
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

func (s *Server) handleMsg(sess *session, msg *ClientMsg) {
	switch msg.Type {
	case MsgTypeFindMatch:
		if sess.isSubscribedToGame() {
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
