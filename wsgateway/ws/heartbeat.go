package ws

import (
	"time"
)

func (s *Server) checkHeartbeat() {
	tick := time.NewTicker(time.Minute)

	for t := range tick.C {
		deadSessions := []*session{}
		s.connsMux.RLock()
		for _, sess := range s.sessions {
			if sess.lastHeartBeat.Load().Before(t.Add(-s.cfg.PingInterval)) {
				deadSessions = append(deadSessions, sess)

			}

		}
		s.connsMux.RUnlock()

		s.stopSessions(true, deadSessions...)
	}
}
