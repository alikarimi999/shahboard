package ws

import (
	"sync"

	"github.com/alikarimi999/shahboard/types"
)

type sessionsManager struct {
	mu       sync.RWMutex
	sessions map[types.ObjectId]map[types.ObjectId]*session // map by userId and sessionId
}

func newSessionsManager() *sessionsManager {
	return &sessionsManager{
		sessions: make(map[types.ObjectId]map[types.ObjectId]*session),
	}
}

func (m *sessionsManager) add(s *session) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.sessions[s.userId]; !ok {
		m.sessions[s.userId] = make(map[types.ObjectId]*session)
	}
	m.sessions[s.userId][s.id] = s
}

func (m *sessionsManager) remove(ss ...*session) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, s := range ss {
		delete(m.sessions[s.userId], s.id)
	}
}

func (m *sessionsManager) getAll() []*session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessions := make([]*session, 0)
	for _, sessionsMap := range m.sessions {
		for _, s := range sessionsMap {
			sessions = append(sessions, s)
		}
	}
	return sessions
}
