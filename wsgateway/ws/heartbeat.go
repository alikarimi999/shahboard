package ws

import (
	"time"
)

func (s *WSServer) checkHeartbeat() {
	tick := time.NewTicker(time.Minute)

	for t := range tick.C {
		deadSessions := []*session{}
		s.connsMux.RLock()
		for _, cs := range s.sessions {
			for _, c := range cs {
				if c.lastHeartBeat.Load().Before(t.Add(-s.cfg.PingInterval)) {
					deadSessions = append(deadSessions, c)

				}
			}
		}
		s.connsMux.RUnlock()

		for _, c := range deadSessions {
			s.stopSession(c)
		}
	}
}
