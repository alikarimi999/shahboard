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

// WSServer is an implementation of entity.StreamIn
type WSServer struct {
	cfg *WsConfigs

	connsMux sync.RWMutex
	sessions map[types.ObjectId][]*session
	counter  *atomic.Int64

	stopCh chan struct{}
	l      log.Logger
}

func NewWSServer(e *gin.Engine, s event.Subscriber, p event.Publisher,
	cfg *WsConfigs, l log.Logger) (*WSServer, error) {
	if cfg == nil {
		cfg = defaultConfigs
	}

	server := &WSServer{
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

func (s *WSServer) addSession(sess *session, userId types.ObjectId) {
	s.connsMux.Lock()
	defer s.connsMux.Unlock()

	s.sessions[userId] = append(s.sessions[userId], sess)
	s.counter.Add(1)
}

func (s *WSServer) removeSession(sess *session) {
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

func (s *WSServer) stopSession(sess ...*session) {
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

func (s *WSServer) isUserPlaying(userId types.ObjectId) bool {
	s.connsMux.Lock()
	defer s.connsMux.Unlock()
	for _, cs := range s.sessions[userId] {
		if cs.isPlaying() {
			return true
		}
	}
	return false
}

func (s *WSServer) sessionReader(sess *session) {
	for {
		_, recievedMsg, err := sess.ReadMessage()
		if err != nil {
			s.l.Debug(fmt.Sprintf("session '%d' read error: %s", sess.id, err))
			return
		}
		var msg clientMsg
		if err := json.Unmarshal(recievedMsg, &msg); err != nil {
			sess.sendErr(0, "invalid message")
			continue
		}

		s.handleMsg(sess, &msg)
	}
}

func (s *WSServer) sessionWriter(sess *session) {
	for {
		select {
		case <-sess.stopCh:
			return
		case message := <-sess.msgCh:
			if err := sess.WriteMessage(websocket.TextMessage, message); err != nil {
				s.stopSession(sess)
			}
		}
	}
}

func (s *WSServer) handleMsg(sess *session, msg *clientMsg) {
	switch msg.Type {
	case msgTypePlay:
		if s.isUserPlaying(sess.userId) {
			sess.sendErr(msg.ID, "already playing a game")
			return
		}
		if sess.gameSubscribed() != 0 {
			sess.sendErr(msg.ID, "already subscribed to a game")
			return
		}
	case msgTypeView:

		var data viewCmdData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			sess.sendErr(msg.ID, "invalid data")
			return
		}

		sess.viewGame(msg.ID, data.GameId)
	case msgTypePing:
		t := time.Now()
		sess.lastHeartBeat.Store(t)
		resp := serverMsg{
			msgBase: msgBase{
				ID:        msg.ID,
				Timestamp: t.Unix(),
				Type:      msgTypePong,
			}}
		sess.send(resp)

	case msgTypeData:

		// handle data message
	}
}
