package user

import (
	"context"
	"fmt"
	"time"

	"github.com/alikarimi999/shahboard/event"
	"github.com/alikarimi999/shahboard/pkg/log"
	"github.com/alikarimi999/shahboard/profileservice/entity"
	"github.com/alikarimi999/shahboard/types"
)

type Repository interface {
	// return nil,nil if user not found
	GetByID(ctx context.Context, id types.ObjectId) (*entity.UserInfo, error)
	Create(ctx context.Context, user *entity.UserInfo) error
	Update(ctx context.Context, userId types.ObjectId, req UpdateUserRequest) error

	// UpdateNX inserts a new user if they don't exist; otherwise, it updates their profile fields.
	// It ensures atomicity using a transaction, committing only after a successful operation.
	UpdateNX(ctx context.Context, user types.ObjectId, email string, req UpdateUserRequest) error
	UpdateLastActiveAt(ctx context.Context, id types.ObjectId, lastActiveAt time.Time) error
}

type RatingService interface {
	GetUserRating(ctx context.Context, userId types.ObjectId) (*entity.Rating, error)
}

type Config struct {
}

type Service struct {
	cfg  Config
	repo Repository
	sub  event.Subscriber
	sm   *event.SubscriptionManager
	rs   RatingService
	l    log.Logger
}

func NewService(cfg Config, repo Repository, sub event.Subscriber, rs RatingService, l log.Logger) *Service {
	s := &Service{
		cfg:  cfg,
		repo: repo,
		sub:  sub,
		rs:   rs,
		l:    l,
	}

	s.sm = event.NewManager(l, s.handleEvent)
	s.sm.AddSubscription(s.sub.Subscribe(event.TopicUser))
	return s
}

func (s *Service) GetUserInfo(ctx context.Context, id types.ObjectId) (*entity.UserInfo, *entity.Rating, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if u == nil {
		return nil, nil, nil
	}

	r, err := s.rs.GetUserRating(ctx, id)
	if err != nil || r == nil {
		s.l.Error(fmt.Sprintf("failed to get user rating: %v", err))
		return u, nil, nil
	}

	return u, r, nil
}

func (s *Service) UpdateUser(ctx context.Context, u types.User, req UpdateUserRequest) error {
	if err := s.repo.UpdateNX(ctx, u.ID, u.Email, req); err != nil {
		s.l.Error(fmt.Sprintf("failed to update user '%s' profile: %v", u.Email, err))
		return err
	}
	s.l.Debug(fmt.Sprintf("user updated profile with UID '%s' Email: '%s'", u.ID, u.Email))

	return nil
}

func (s *Service) handleEvent(e event.Event) {
	t := e.GetTopic()
	switch t.Domain() {
	case event.DomainUser:
		switch t.Action() {
		case event.ActionCreated:
			s.handleUserCreated(e.(*event.EventUserCreated))
		case event.ActionLoggedIn:
			s.handleUserLoggedIn(e.(*event.EventUserLoggedIn))
		}
	}
}

func (s *Service) handleUserCreated(e *event.EventUserCreated) {
	var u *entity.UserInfo
	if e.IsGuest {
		u = wrapUserInfoForGuest(e)
	} else {
		t := time.Now()
		u = &entity.UserInfo{
			ID:           e.UserID,
			Email:        e.Email,
			Name:         e.Name,
			AvatarUrl:    e.Picture,
			CreatedAt:    t,
			LastActiveAt: t,
		}
	}

	if err := s.repo.Create(context.Background(), u); err != nil {
		s.l.Error(err.Error())
		return
	}
	s.l.Debug(fmt.Sprintf("user created with UID: '%s' Email: '%s'", e.UserID, e.Email))
}

func wrapUserInfoForGuest(e *event.EventUserCreated) *entity.UserInfo {
	t := time.Now()

	return &entity.UserInfo{
		ID:           e.UserID,
		Email:        e.Email,
		Name:         fmt.Sprintf("guest_%s", e.UserID.String()),
		CreatedAt:    t,
		LastActiveAt: t,
	}

}

func (s *Service) handleUserLoggedIn(e *event.EventUserLoggedIn) {
	if err := s.repo.UpdateLastActiveAt(context.Background(), e.UserID, time.Now()); err != nil {
		s.l.Error(err.Error())
		return
	}

	s.l.Debug(fmt.Sprintf("user logged in with UID: '%s' Email: '%s'", e.UserID, e.Email))
}
