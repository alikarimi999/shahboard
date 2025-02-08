package ws

import (
	"time"
)

func (s *Server) checkHeartbeat() {
	tick := time.NewTicker(time.Second * 5)

	for t := range tick.C {
		deadSessions := []*session{}
		disconnectedSessions := []*session{}
		s.connsMux.RLock()
		for _, sess := range s.sessions {
			lh := sess.lastHeartBeat.Load()
			if lh.Before(t.Add(-s.cfg.PingIntervalDeadSession)) {
				deadSessions = append(deadSessions, sess)
				continue
			}
			if sess.isClosed() {
				continue
			}
			if lh.Before(t.Add(-s.cfg.PingIntervalDisconnectedSession)) {
				disconnectedSessions = append(disconnectedSessions, sess)
			}
		}
		s.connsMux.RUnlock()

		if len(deadSessions) > 0 {
			s.stopSessions(true, deadSessions...)
		}
		if len(disconnectedSessions) > 0 {
			s.stopSessions(false, disconnectedSessions...)
		}
	}

}
