package ws

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

func (s *session) readLoop() {
	// defer func() {
	// 	s.l.Debug(fmt.Sprintf("session '%s' reader stopped", se.id))
	// }()
	s.wg.Add(1)
	defer s.wg.Done()
	for {
		mt, recievedMsg, err := s.ReadMessage()
		if err != nil {
			// s.l.Debug(fmt.Sprintf("session '%s' read message error: %v", se.id, err))
			s.Stop()
			return
		}

		if mt == websocket.BinaryMessage && len(recievedMsg) > 0 && recievedMsg[0] == 0x0 {
			s.lastHeartBeat.Store(time.Now())
			s.sendPong()
			continue
		}

		if mt == websocket.TextMessage && len(recievedMsg) > 0 {
			var msg Msg
			if err := json.Unmarshal(recievedMsg, &msg); err != nil {
				continue
			}
			s.handleMsg(s, &msg)
		}
	}
}

func (s *session) writeLoop() {
	// defer func() {
	// 	s.l.Debug(fmt.Sprintf("session '%s' writer stopped", se.id))
	// }()
	s.wg.Add(1)
	defer s.wg.Done()
	for {
		select {
		case <-s.stopCh:
			return
		case message := <-s.msgCh:
			if message == nil {
				return
			}

			if err := s.WriteMessage(websocket.TextMessage, message); err != nil {
				// s.l.Debug(fmt.Sprintf("session '%s' write message error: %v", se.id, err))
				s.Stop()
				return
			}
		case <-s.pongCh:
			if err := s.WriteMessage(websocket.BinaryMessage, []byte{0x1}); err != nil {
				s.l.Debug(fmt.Sprintf("session '%s' write pong message error: %v", s.id, err))
				s.Stop()
				return
			}
		}
	}
}
