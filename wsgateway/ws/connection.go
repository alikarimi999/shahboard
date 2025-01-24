package ws

import (
	"encoding/json"
	"time"

	"github.com/alikarimi999/shahboard/event"
	"github.com/alikarimi999/shahboard/types"
	"github.com/gorilla/websocket"
	"go.uber.org/atomic"
)

const (
	gameViewerRole uint8 = iota + 1
	gamePlayerRole
)

type gameSubscription struct {
	gameID types.ObjectId
	role   uint8
}

type session struct {
	*websocket.Conn
	id     string
	userId types.ObjectId
	msgCh  chan []byte

	game gameSubscription

	lastHeartBeat *atomic.Time

	s event.Subscriber
	p event.Publisher

	closed *atomic.Bool
	stopCh chan struct{}
	pongCh chan string
}

func (s *session) send(msg serverMsg) {
	b, _ := json.Marshal(msg)
	s.msgCh <- b
}

func (s *session) gameSubscribe(gameId types.ObjectId, role uint8) {
	s.game = gameSubscription{
		gameID: gameId,
		role:   role,
	}
}

func (s *session) gameSubscribed() types.ObjectId {
	return s.game.gameID
}

func (s *session) gameUnsubscribe() {
	s.game = gameSubscription{}
}

func (s *session) isPlaying() bool {
	return s.game.role == gamePlayerRole
}

func (s *session) sendErr(id types.ObjectId, err string) {
	s.send(serverMsg{
		msgBase: msgBase{
			ID:        id,
			Type:      msgTypeErr,
			Timestamp: time.Now().Unix(),
		},
		Data: []byte(err),
	})
}
