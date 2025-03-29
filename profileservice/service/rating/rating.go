package rating

import (
	"context"
	"fmt"
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
	// update ratings and game elo changes atomically
	Update(ctx context.Context, ratings []*entity.Rating, changes []*entity.GameEloChange) error

	GetGameEloChangesByUserId(ctx context.Context, userId types.ObjectId) ([]*entity.GameEloChange, error)
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

func (s *Service) GetUserChangeHistory(ctx context.Context, userId types.ObjectId) ([]*entity.GameEloChange, error) {
	return s.repo.GetGameEloChangesByUserId(ctx, userId)
}

func (s *Service) handleEvent(e event.Event) {
	switch e.GetTopic().Domain() {
	case event.DomainGame:
		switch e.GetTopic().Action() {
		case event.ActionEnded:
			s.handleGameEnded(e.(*event.EventGameEnded))
		}
	}
}

func (s *Service) handleGameEnded(e *event.EventGameEnded) {
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

	s1 := calcScore1(e.Outcome, e.Player1.Color)
	elo1 := elo.CalculateElo(r1.CurrentScore, r2.CurrentScore, s1)
	elo2 := elo.CalculateElo(r2.CurrentScore, r1.CurrentScore, 1-s1)
	t := time.Now()

	var result1, result2 entity.GameResult
	if s1 == 1 {
		result1 = entity.GameResultWin
		result2 = entity.GameResultLoss
	} else if s1 == 0 {
		result1 = entity.GameResultLoss
		result2 = entity.GameResultWin
	} else {
		result1 = entity.GameResultDraw
		result2 = entity.GameResultDraw
	}

	c1 := &entity.GameEloChange{
		UserId:     e.Player1.ID,
		EloChange:  elo1 - r1.CurrentScore,
		GameId:     e.GameID,
		OpponentId: e.Player2.ID,
		Result:     result1,
		UpdatedAt:  t,
	}
	c2 := &entity.GameEloChange{
		UserId:     e.Player2.ID,
		EloChange:  elo2 - r2.CurrentScore,
		GameId:     e.GameID,
		OpponentId: e.Player1.ID,
		Result:     result2,
		UpdatedAt:  t,
	}

	r1.CurrentScore = elo1
	r2.CurrentScore = elo2

	r1.LastUpdated = t
	r2.LastUpdated = t

	if r1.BestScore < elo1 {
		r1.BestScore = elo1
	}

	if r2.BestScore < elo2 {
		r2.BestScore = elo2
	}

	updateGameStats(r1, r2, e.Outcome, e.Player1.Color)

	if err := s.repo.Update(ctx, []*entity.Rating{r1, r2}, []*entity.GameEloChange{c1, c2}); err != nil {
		// TODO: handle this situation better
		s.l.Error(err.Error())
		return
	}

	s.l.Debug(fmt.Sprintf("Game '%s' ended, players ratings updated", e.GameID))
}

func calcScore1(o types.GameOutcome, p1Color types.Color) float64 {
	switch o {
	case types.WhiteWon:
		if p1Color == types.ColorWhite {
			return 1
		}
		return 0
	case types.BlackWon:
		if p1Color == types.ColorBlack {
			return 1
		}
		return 0
	default:
		return 0.5

	}
}

func updateGameStats(r1, r2 *entity.Rating, outcome types.GameOutcome, p1Color types.Color) {
	r1.GamesPlayed++
	r2.GamesPlayed++

	switch outcome {
	case types.WhiteWon:
		winner, loser := r1, r2
		if p1Color != types.ColorWhite {
			winner, loser = r2, r1
		}
		winner.GamesWon++
		loser.GamesLost++

	case types.BlackWon:
		winner, loser := r1, r2
		if p1Color != types.ColorBlack {
			winner, loser = r2, r1
		}
		winner.GamesWon++
		loser.GamesLost++

	case types.Draw:
		r1.GamesDraw++
		r2.GamesDraw++
	}
}
