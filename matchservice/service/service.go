package match

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/alikarimi999/shahboard/event"
	"github.com/alikarimi999/shahboard/pkg/elo"
	"github.com/alikarimi999/shahboard/pkg/log"
	"github.com/alikarimi999/shahboard/types"
)

type Config struct {
	EngineTicker       int `json:"engine_ticker"`
	MatchRequestTicker int `json:"match_request_ticker"`
}

type Service struct {
	cfg Config
	e   *engine

	p      event.Publisher
	rating RatingService
	game   GameService

	stopCh chan struct{}
	wg     sync.WaitGroup

	l log.Logger
}

func NewService(cfg Config, p event.Publisher, score RatingService, game GameService, l log.Logger) (*Service, error) {
	s := &Service{
		cfg:    cfg,
		e:      newEngine(time.Duration(cfg.EngineTicker) * time.Second),
		p:      p,
		rating: score,
		game:   game,
		stopCh: make(chan struct{}),
		l:      l,
	}

	s.run()

	return s, nil
}

func (s *Service) NewMatchRequest(ctx context.Context, userId types.ObjectId) (*event.EventUsersMatchCreated, error) {
	t := time.NewTicker(time.Duration(s.cfg.MatchRequestTicker) * time.Second)

	currentGameId, err := s.game.GetUserLiveGameID(ctx, userId)
	if err != nil {
		s.l.Error(fmt.Sprintf("failed to get user '%s' live game id: %s", userId, err.Error()))
		return nil, fmt.Errorf("internal error")
	}

	if currentGameId != types.ObjectZero {
		s.l.Debug(fmt.Sprintf("user '%s' is already in a game", userId))
		return nil, fmt.Errorf("user is already in a game")
	}

	score, err := s.rating.GetUserScore(userId)
	if err != nil {
		s.l.Error(fmt.Sprintf("failed to get user '%s' level: %s", userId, err.Error()))
		score = elo.BaseScore
	}

	req, ok := s.e.addToQueue(userId, score)
	if !ok {
		return nil, fmt.Errorf("user '%s' already has a match request", userId)
	}

	// s.l.Debug(fmt.Sprintf("New match request for user '%s' with level %d", userId, score))
	select {
	case <-ctx.Done():
		s.e.cancelRequest(req)
		return nil, ctx.Err()
	case res := <-req.response():
		return res, nil
	case <-t.C:
		s.e.cancelRequest(req)
		return nil, fmt.Errorf("request timeout")
	}

}

func (s *Service) run() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case ms := <-s.e.listen():
				if len(ms) == 0 {
					continue
				}
				events := make([]event.Event, 0, len(ms))
				for _, m := range ms {
					s.l.Debug(fmt.Sprintf("match '%s' for user '%s' and '%s'", m.ID, m.User1.ID, m.User2.ID))
					events = append(events, m)
				}

				if err := s.p.Publish(events...); err != nil {
					s.l.Error(err.Error())
				}

			case <-s.stopCh:
				s.e.stop()

				return
			}
		}
	}()
}

func (s *Service) Stop() {
	close(s.stopCh)
	s.wg.Wait()
}
