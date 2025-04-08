package bot

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/alikarimi999/shahboard/wsgateway/ws"
	"github.com/gorilla/websocket"
)

type webSocket struct {
	b      *Bot
	conn   *websocket.Conn
	once   sync.Once
	stopCh chan struct{}
}

func (b *Bot) SetupWS() error {
	conn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("%s/wsgateway/ws?token=%s", b.url, b.jwtToken), nil)
	if err != nil {
		return err
	}
	b.ws = &webSocket{
		b:      b,
		conn:   conn,
		stopCh: make(chan struct{})}

	b.ws.waitFor(ws.MsgTypeWelcome)
	go b.ws.handlePingPong()
	go b.ws.runReader()

	return nil
}

func (b *Bot) SendWsMessage(msg ws.Msg) error {
	return b.ws.sendMessage(msg)
}

func (s *webSocket) waitFor(t ws.MsgType) ws.Msg {
	msg := ws.Msg{}
	for {
		if err := s.conn.ReadJSON(&msg); err != nil {
			if msg.Type == t {
				return msg
			}
		}
	}
}

func (s *webSocket) stop() {
	s.once.Do(func() {
		s.stopCh <- struct{}{}
		s.conn.Close()
	})
}

func (s *webSocket) sendMessage(msg ws.Msg) error {
	return s.conn.WriteJSON(msg)
}

func (s *webSocket) handlePingPong() {
	for {
		select {
		case <-s.stopCh:
			return
		default:
			if err := s.conn.WriteMessage(websocket.BinaryMessage, []byte{0x0}); err != nil {
				return
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func (s *webSocket) runReader() {
	for {
		select {
		case <-s.stopCh:
			return
		default:
			mt, b, err := s.conn.ReadMessage()
			if err != nil {
				s.stop()
				return
			}
			if mt == websocket.TextMessage {
				msg := &ws.Msg{}
				if err := json.Unmarshal(b, msg); err != nil {
					s.stop()
					return
				}
				s.b.Publish(Event{
					Topic: Topic(msg.Type),
					Data:  msg,
				})
			}
		}
	}
}
