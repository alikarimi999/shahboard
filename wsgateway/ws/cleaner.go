package ws

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/alikarimi999/shahboard/event"
)

// sessionCleaner starts a background goroutine that buffers incoming session cleanup requests
// from the cleanerCh channel and processes them in batches based on a specified batch size or flush interval.
// this is not a real Redis-Level batch cleaner, and to turn into a real batch cleaner,
// i need redis cache methods that support batch operations by piplining or lua scripts.
type sessionCleaner struct {
	s         *Server
	cleanerCh chan *session

	batchSize     int
	flushInterval time.Duration
}

func newSessionCleaner(s *Server, batchSize int, flushInterval time.Duration) *sessionCleaner {
	c := &sessionCleaner{
		s:             s,
		cleanerCh:     make(chan *session, 1000),
		batchSize:     batchSize,
		flushInterval: flushInterval,
	}
	c.run()

	return c
}

func (c *sessionCleaner) run() {
	go func() {
		var (
			batch []*session
			timer = time.NewTimer(c.flushInterval)
			ctx   = context.Background()
		)

		defer timer.Stop()

		for {
			select {
			case sess, ok := <-c.cleanerCh:
				if !ok {
					c.removeSessionsInBatch(ctx, batch)
					return
				}
				batch = append(batch, sess)
				if len(batch) >= c.batchSize {
					c.removeSessionsInBatch(ctx, batch)
					batch = nil
					safeResetTimer(timer, c.flushInterval)
				}
			case <-timer.C:
				if len(batch) > 0 {
					c.removeSessionsInBatch(ctx, batch)
					batch = nil
				}
				timer.Reset(c.flushInterval)
			}
		}
	}()
}

func (c *sessionCleaner) clean(sess *session) {
	c.cleanerCh <- sess
}

// removeSessionsInBatch performs session cleanup concurrently for a batch of sessions.
// It's not a real Redis-Level batch cleaner,
// and it's just a temp solution until I implement redis methods for batch operation.
func (c *sessionCleaner) removeSessionsInBatch(ctx context.Context, sesstions []*session) {
	counter := 0
	sem := make(chan struct{}, c.batchSize)
	wg := &sync.WaitGroup{}

	for _, sess := range sesstions {
		counter++
		wg.Add(1)
		sem <- struct{}{}

		go func() {
			defer func() {
				wg.Done()
				<-sem
			}()

			c.s.removeSession(ctx, sess)
		}()
	}
	wg.Wait()

	if counter >= c.batchSize {
		c.s.l.Info(fmt.Sprintf("a batch of '%d' sessions cleaned from cache", counter))
	}
}

func (s *Server) removeSession(ctx context.Context, se *session) {
	s.sm.remove(se.userId, se.id)

	ids := se.getAllViewGames()
	if len(ids) > 0 {
		if err := s.cache.removeFromGameViewersList(ctx, se.userId, ids...); err != nil {
			s.l.Error(err.Error())
		}
	}

	playGameId := se.playGameId.Load()
	if !playGameId.IsZero() {
		counter, err := s.cache.deleteUserGameSession(ctx, se.userId, se.id)
		if err != nil {
			// TODO: need to handle this error
			s.l.Error(err.Error())

		} else if counter == 0 {
			// this is the last session for this user that is playing the game
			// so we need to notify other parts of system

			if err := s.p.Publish(event.EventGamePlayerLeft{
				GameID:    playGameId,
				PlayerID:  se.userId,
				Timestamp: time.Now().Unix(),
			}); err != nil {
				s.l.Error(err.Error())
			}
		}
	} else {
		if err := s.cache.deleteUserSession(ctx, se.userId, se.id); err != nil {
			s.l.Error(err.Error())
		}
	}
}

func safeResetTimer(t *time.Timer, d time.Duration) {
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
	t.Reset(d)
}
