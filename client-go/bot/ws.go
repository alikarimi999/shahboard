package bot

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/alikarimi999/shahboard/wsgateway/ws"
	"github.com/gorilla/websocket"
)

type webSocket struct {
	b    *Bot
	conn *websocket.Conn
	once sync.Once

	writeCh chan ws.Msg

	stopCh chan struct{}
}

func (b *Bot) SetupWS() error {
	u, err := url.Parse(b.url)
	if err != nil {
		return err
	}

	wsURL := fmt.Sprintf("%s://%s", "ws", u.Host)
	conn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("%s/wsgateway/ws?token=%s", wsURL, b.jwtToken), nil)
	if err != nil {
		return err
	}
	b.ws = &webSocket{
		b:       b,
		conn:    conn,
		writeCh: make(chan ws.Msg, 100),
		stopCh:  make(chan struct{})}

	// b.ws.waitFor(ws.MsgTypeWelcome)

	go b.ws.runWriter()
	go b.ws.runReader()

	s := b.Subscribe(Topic(ws.MsgTypeWelcome))
	defer s.Unsubscribe()
	for {
		select {
		case <-s.Consume():
			// fmt.Printf("bot %s received welcome message\n", b.Email())
			return nil
		case <-time.After(60 * time.Second):
			return fmt.Errorf("bot %s didn't receiv welcome message", b.Email())
		}
	}
}

func (b *Bot) SendWsMessage(msg ws.Msg) error {
	return b.ws.sendMessage(msg)
}

func (s *webSocket) stop() {
	s.once.Do(func() {
		s.stopCh <- struct{}{}
		s.conn.Close()
	})
}

func (s *webSocket) sendMessage(msg ws.Msg) error {
	select {
	case s.writeCh <- msg:
		return nil
	default:
		return fmt.Errorf("failed to send message")
	}
}

func (s *webSocket) runWriter() {
	t := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-s.stopCh:
			return

		case msg := <-s.writeCh:
			if err := s.conn.WriteJSON(msg); err != nil {
				fmt.Printf("bot '%s' error sending ping message: %v\n", s.b.Email(), err)
				s.stop()
				return
			}
		case <-t.C:
			if err := s.conn.WriteMessage(websocket.BinaryMessage, []byte{0x0}); err != nil {
				fmt.Printf("bot '%s' error sending ping message: %v\n", s.b.Email(), err)
				s.stop()
				return
			}
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
				fmt.Printf("bot '%s' error reading message: %v\n", s.b.Email(), err)
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
