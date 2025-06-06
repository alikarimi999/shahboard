package match

import (
	"sync"
	"time"

	"github.com/alikarimi999/shahboard/event"
	"github.com/alikarimi999/shahboard/pkg/elo"
	"github.com/alikarimi999/shahboard/types"
)

type engine struct {
	t     time.Ticker
	mu    sync.Mutex
	queue map[types.Level][]*matchRequest

	matchCh chan []*event.EventUsersMatchCreated
	stopCh  chan struct{}
	wg      sync.WaitGroup
}

func newEngine(ticker time.Duration) *engine {
	e := &engine{
		t:       *time.NewTicker(ticker),
		queue:   make(map[types.Level][]*matchRequest, 0),
		matchCh: make(chan []*event.EventUsersMatchCreated),
		stopCh:  make(chan struct{}),
	}

	e.run()

	return e
}

func (e *engine) run() {
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		for {
			select {
			case <-e.t.C:
				e.matchCh <- e.findMatches()
			case <-e.stopCh:
				return
			}
		}

	}()
}

func (e *engine) listen() <-chan []*event.EventUsersMatchCreated { return e.matchCh }

func (e *engine) stop() {
	close(e.stopCh)
	e.wg.Wait()
}

func (e *engine) addToQueue(pId types.ObjectId, s int64) (*matchRequest, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, q := range e.queue {
		for _, req := range q {
			if req.userId == pId {
				return nil, false
			}
		}
	}

	l := elo.GetPlayerLevel(s)
	r := newMatchRequest(pId, s, l)
	e.queue[l] = append(e.queue[l], r)

	return r, true
}

func (e *engine) cancelRequest(r *matchRequest) {
	e.mu.Lock()
	defer e.mu.Unlock()

	users := e.queue[r.level]
	for i, req := range users {
		if req.userId == r.userId {
			e.queue[r.level] = append(users[:i], users[i+1:]...)
			return
		}
	}
}

func (e *engine) findMatches() []*event.EventUsersMatchCreated {
	e.mu.Lock()
	defer e.mu.Unlock()

	var leftover []*matchRequest
	var matches []*event.EventUsersMatchCreated

	t := time.Now().Unix()
	for l := types.LevelKing; l >= types.LevelPawn; l-- {
		currentQueue := append(leftover, e.queue[l]...)
		leftover = nil

		for len(currentQueue) > 1 {
			u1 := currentQueue[0]
			u2 := currentQueue[1]
			m := &event.EventUsersMatchCreated{
				ID:        types.NewObjectId(),
				User1:     types.User{ID: u1.userId, Score: u1.score},
				User2:     types.User{ID: u2.userId, Score: u2.score},
				Timestamp: t,
			}

			u1.sendResponse(m)
			u2.sendResponse(m)

			matches = append(matches, m)
			currentQueue = currentQueue[2:]
		}

		if len(currentQueue) > 0 {
			leftover = append(leftover, currentQueue...)
		}

		e.queue[l] = nil

		if l == types.LevelPawn {
			for _, r := range leftover {
				e.queue[r.level] = append(e.queue[r.level], r)
			}
		}
	}

	return matches
}

type matchRequest struct {
	userId types.ObjectId
	score  int64
	level  types.Level
	ch     chan *event.EventUsersMatchCreated
}

func newMatchRequest(pId types.ObjectId, s int64, l types.Level) *matchRequest {
	return &matchRequest{
		userId: pId,
		score:  s,
		level:  l,
		ch:     make(chan *event.EventUsersMatchCreated, 1),
	}
}

func (m matchRequest) sendResponse(r *event.EventUsersMatchCreated) {
	m.ch <- r
	close(m.ch)
}

func (m *matchRequest) response() <-chan *event.EventUsersMatchCreated { return m.ch }
