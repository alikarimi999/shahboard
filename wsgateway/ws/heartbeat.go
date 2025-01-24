package ws

import (
	"time"
)

func (s *WSServer) checkHeartbeat() {
	tick := time.NewTicker(time.Minute)

	for t := range tick.C {
		connections := []*session{}
		s.connsMux.RLock()
		for _, cs := range s.sessions {
			for _, c := range cs {
				if c.lastHeartBeat.Load().Before(t.Add(-s.cfg.PingInterval)) {
					connections = append(connections, c)

				}
			}
		}
		s.connsMux.RUnlock()

		for _, c := range connections {
			s.stopConnection(c)
		}
	}
}
