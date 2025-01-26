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
		user := u.(types.User)
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

	})

	return server, nil
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

func (s *Server) stopSession(sess ...*session) {
	for _, se := range sess {
		if se.closed.Load() {
			continue
		}
		se.closed.Store(true)
		close(se.stopCh)
		se.Close()
		s.removeSession(se)
	}
}

func (s *Server) sessionReader(sess *session) {
	for {
		_, recievedMsg, err := sess.ReadMessage()
		if err != nil {
			s.l.Debug(fmt.Sprintf("session '%d' read error: %s", sess.id, err))
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
	for {
		select {
		case <-sess.stopCh:
			return
		case message := <-sess.msgCh:
			if err := sess.WriteMessage(websocket.TextMessage, message); err != nil {
				s.stopSession(sess)
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
