package ws

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/alikarimi999/shahboard/types"
)

// WSServer is an implementation of entity.StreamIn
type WSServer struct {
	cfg *WsConfigs

	connsMux sync.RWMutex
	sessions map[types.ObjectId][]*session
	counter  *atomic.Int64
}

func (s *WSServer) addConnection(conn *session, userId types.ObjectId) {
	s.connsMux.Lock()
	defer s.connsMux.Unlock()
	conn.id = fmt.Sprintf("%s-%d", userId, len(s.sessions[userId]))
	s.sessions[userId] = append(s.sessions[userId], conn)
	s.counter.Add(1)
}

func (s *WSServer) removeConnection(conn *session) {
	s.connsMux.Lock()
	defer s.connsMux.Unlock()
	for i, c := range s.sessions[conn.userId] {
		if c.id == conn.id {
			s.sessions[conn.userId] = append(s.sessions[conn.userId][:i], s.sessions[conn.userId][i+1:]...)
			s.counter.Add(-1)
			break
		}
	}
}

func (s *WSServer) stopConnection(conns ...*session) {
	for _, conn := range conns {
		if conn.closed.Load() {
			continue
		}
		conn.closed.Store(true)
		close(conn.stopCh)
		conn.Close()
		s.removeConnection(conn)
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

func (s *WSServer) handleMsg(conn *session, msg clientMsg) {
	switch msg.Type {
	case msgTypePlay:
		if s.isUserPlaying(conn.userId) {
			conn.sendErr(msg.ID, "already playing a game")
			return
		}
		if conn.gameSubscribed() != 0 {
			conn.sendErr(msg.ID, "already subscribed to a game")
			return
		}
	case msgTypeView:
		if conn.gameSubscribed() != 0 {
			conn.sendErr(msg.ID, "already subscribed to a game")
			return
		}
		var data viewCmdData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			conn.sendErr(msg.ID, "invalid data")
			return
		}

		conn.gameSubscribe(data.GameId, gameViewerRole)
	case msgTypeData:
		// handle data message
	}
}
