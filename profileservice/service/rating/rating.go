package rating

import (
	"context"
	"time"

	"github.com/alikarimi999/shahboard/event"
	"github.com/alikarimi999/shahboard/pkg/elo"
	"github.com/alikarimi999/shahboard/pkg/log"
	"github.com/alikarimi999/shahboard/profileservice/entity"
	"github.com/alikarimi999/shahboard/types"
)

type Repository interface {
	// return nil if not found
	GetByUserId(ctx context.Context, id types.ObjectId) (*entity.Rating, error)
	Update(ctx context.Context, ratings ...*entity.Rating) error
}

type Config struct {
}

type Service struct {
	cfg  Config
	repo Repository
	sub  event.Subscriber
	sm   *event.SubscriptionManager
	l    log.Logger
}

// implement user service and rating service in one service for simplicity and faster development
func NewService(cfg Config, repo Repository, sub event.Subscriber, l log.Logger) *Service {
	s := &Service{
		cfg:  cfg,
		repo: repo,
		sub:  sub,
		l:    l,
	}

	s.sm = event.NewManager(l, s.handleEvent)
	s.sm.AddSubscription(s.sub.Subscribe(event.TopicGame))

	return s
}

// If user not found, create a new rating for the user with base rating
func (s *Service) GetUserRating(ctx context.Context, userId types.ObjectId) (*entity.Rating, error) {
	r, err := s.repo.GetByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}

	if r == nil {
		r = &entity.Rating{
			UserId:       userId,
			CurrentScore: elo.BaseRating,
		}
	}

	return r, nil
}

func (s *Service) handleEvent(e event.Event) {
	t := e.GetTopic()
	switch t {
	case event.TopicGame:
		switch t.Action() {
		case event.ActionEnded:
			s.handleGameEnded(e.(event.EventGameEnded))
		}
	}
}

func (s *Service) handleGameEnded(e event.EventGameEnded) {
	ctx := context.Background()
	r1, err := s.repo.GetByUserId(ctx, e.Player1.ID)
	if err != nil {
		// TODO: handle this situation better
		s.l.Error(err.Error())
		return
	}

	if r1 == nil {
		r1 = &entity.Rating{
			UserId:       e.Player1.ID,
			CurrentScore: elo.BaseRating,
		}
	}

	r2, err := s.repo.GetByUserId(ctx, e.Player2.ID)
	if err != nil {
		// TODO: handle this situation better
		s.l.Error(err.Error())
		return
	}

	if r2 == nil {
		r2 = &entity.Rating{
			UserId:       e.Player2.ID,
			CurrentScore: elo.BaseRating,
		}
	}

	s1 := calcScore1(e.Outcome)
	elo1 := elo.CalculateElo(r1.CurrentScore, r2.CurrentScore, s1)
	elo2 := elo.CalculateElo(r2.CurrentScore, r1.CurrentScore, 1-s1)

	r1.CurrentScore = elo1
	r2.CurrentScore = elo2

	t := time.Now()
	r1.LastUpdated = t
	r2.LastUpdated = t

	if r1.BestScore < elo1 {
		r1.BestScore = elo1
	}

	if r2.BestScore < elo2 {
		r2.BestScore = elo2
	}

	updateGameStats(r1, r2, e.Outcome)

	if err := s.repo.Update(ctx, r1, r2); err != nil {
		// TODO: handle this situation better
		s.l.Error(err.Error())
	}
}

func calcScore1(o types.GameOutcome) float64 {
	switch o {
	case types.WhiteWon:
		return 1
	case types.BlackWon:
		return 0
	default:
		return 0.5

	}
}

func updateGameStats(r1, r2 *entity.Rating, outcome types.GameOutcome) {
	r1.GamesPlayed++
	r2.GamesPlayed++
	if outcome == types.WhiteWon {
		r1.GamesWon++
		r2.GamesLost++
	} else if outcome == types.BlackWon {
		r1.GamesLost++
		r2.GamesWon++
	} else {
		r1.GamesDraw++
		r2.GamesDraw++
	}
}
