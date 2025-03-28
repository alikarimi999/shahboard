package user

import (
	"context"
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
	Update(ctx context.Context, user *entity.UserInfo) error
	UpdateLastActiveAt(ctx context.Context, id types.ObjectId, lastActiveAt time.Time) error
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

func NewService(cfg Config, repo Repository, sub event.Subscriber, l log.Logger) *Service {
	s := &Service{
		cfg:  cfg,
		repo: repo,
		sub:  sub,
		l:    l,
	}

	s.sm = event.NewManager(l, s.handleEvent)
	s.sm.AddSubscription(s.sub.Subscribe(event.TopicUser))
	return s
}

func (s *Service) GetUserInfo(ctx context.Context, id types.ObjectId) (*entity.UserInfo, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) UpdateUser(ctx context.Context, user *entity.UserInfo) error {
	return s.repo.Update(ctx, user)
}

func (s *Service) handleEvent(e event.Event) {
	t := e.GetTopic()
	switch t.Domain() {
	case event.DomainUser:
		switch t.Action() {
		case event.ActionCreated:
			s.handleUserCreated(e.(event.EventUserCreated))
		case event.ActionLoggedIn:
			s.handleUserLoggedIn(e.(event.EventUserLoggedIn))
		}
	}
}

func (s *Service) handleUserCreated(e event.EventUserCreated) {
	t := time.Now()
	if err := s.repo.Create(context.Background(), &entity.UserInfo{
		ID:           e.UserID,
		Email:        e.Email,
		Name:         e.Name,
		AvatarUrl:    e.Picture,
		CreatedAt:    t,
		LastActiveAt: t,
	}); err != nil {
		s.l.Error(err.Error())
	}
}

func (s *Service) handleUserLoggedIn(e event.EventUserLoggedIn) {
	if err := s.repo.UpdateLastActiveAt(context.Background(), e.UserID, time.Now()); err != nil {
		s.l.Error(err.Error())
	}
}
