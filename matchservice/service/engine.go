package match

import (
	"fmt"
	"sync"
	"time"

	"github.com/alikarimi999/shahboard/types"
)

type engine struct {
	t     time.Ticker
	mu    sync.Mutex
	queue map[types.Level][]*matchRequest

	matchCh chan []*Match
	stopCh  chan struct{}
	wg      sync.WaitGroup
}

func newEngine(ticker time.Duration) *engine {
	fmt.Println("ticker ", ticker)
	e := &engine{
		t:       *time.NewTicker(ticker),
		queue:   make(map[types.Level][]*matchRequest, 0),
		matchCh: make(chan []*Match),
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

func (e *engine) listen() <-chan []*Match { return e.matchCh }

func (e *engine) stop() {
	close(e.stopCh)
	e.wg.Wait()
}

func (e *engine) addToQueue(pId types.ObjectId, l types.Level) *matchRequest {
	e.mu.Lock()
	defer e.mu.Unlock()

	r := newMatchRequest(pId, l)
	e.queue[l] = append(e.queue[l], r)

	return r
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

func (e *engine) findMatches() []*Match {
	e.mu.Lock()
	defer e.mu.Unlock()

	var leftover []*matchRequest
	var matches []*Match

	t := time.Now().Unix()
	for l := types.LevelKing; l >= types.LevelPawn; l-- {
		currentQueue := append(leftover, e.queue[l]...)
		leftover = nil

		for len(currentQueue) > 1 {
			u1 := currentQueue[0]
			u2 := currentQueue[1]
			m := &Match{
				ID:        types.NewObjectId(),
				UserA:     types.User{ID: u1.userId, Level: u1.level},
				UserB:     types.User{ID: u2.userId, Level: u2.level},
				TimeStamp: t,
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
	fmt.Println()

	return matches
}

type matchRequest struct {
	userId types.ObjectId
	level  types.Level
	ch     chan *Match
}

func newMatchRequest(pId types.ObjectId, l types.Level) *matchRequest {
	return &matchRequest{
		userId: pId,
		level:  l,
		ch:     make(chan *Match, 1),
	}
}

func (m matchRequest) sendResponse(r *Match) {
	m.ch <- r
	close(m.ch)
}

func (m *matchRequest) response() <-chan *Match { return m.ch }
