package service

import (
	"context"

	"github.com/alikarimi999/shahboard/event"
	"github.com/alikarimi999/shahboard/pkg/log"
	"github.com/alikarimi999/shahboard/types"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	InstanceID string `json:"instance_id"`
}

type Service struct {
	cfg   Config
	cm    *chatsManager
	sm    *event.SubscriptionManager
	cache *redisChatCache

	pub event.Publisher
	sub event.Subscriber

	l log.Logger
}

func NewService(cfg Config, pub event.Publisher, sub event.Subscriber, rc *redis.Client, l log.Logger) *Service {
	s := &Service{
		cfg:   cfg,
		cm:    newChatsManager(),
		cache: newRedisChatCache(cfg.InstanceID, rc),
		pub:   pub,
		sub:   sub,
		l:     l,
	}
	s.sm = event.NewManager(l, s.handleEvents)
	s.sm.AddSubscription(s.sub.Subscribe(event.TopicGameCreated))

	return s
}

func (s *Service) CreateGameChat(ctx context.Context, gameId types.ObjectId, player1, player2 types.Player) (bool, error) {
	gameChat := s.cm.createChat(gameId, player1, player2)
	if gameChat == nil {
		return false, nil
	}

	ok, err := s.cache.addGameChat(ctx, gameChat)
	if err != nil {
		return false, err
	}

	return ok, nil
}
