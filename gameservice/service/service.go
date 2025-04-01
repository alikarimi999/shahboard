package game

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/alikarimi999/shahboard/event"
	"github.com/alikarimi999/shahboard/gameservice/entity"
	"github.com/alikarimi999/shahboard/pkg/log"
	"github.com/alikarimi999/shahboard/types"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	cfg Config

	sm *event.SubscriptionManager

	gMu   sync.RWMutex
	games map[types.ObjectId]*entity.Game

	cache *redisGameCache

	live *liveGamesService

	pub event.Publisher
	sub event.Subscriber

	l log.Logger

	closeCh chan struct{}
}

func NewGameService(cfg Config, redis *redis.Client, pub event.Publisher, sub event.Subscriber,
	ws WsGateway, l log.Logger) (*Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	s := &Service{
		cfg: cfg,

		games: make(map[types.ObjectId]*entity.Game),
		cache: newRedisGameCache(cfg.InstanceID, redis, 15*time.Minute, l),

		pub: pub,
		sub: sub,
		l:   l,

		closeCh: make(chan struct{}),
	}
	s.live = newLiveGamesService(s.cache, ws, l)

	s.sm = event.NewManager(l, s.handleEvents)
	s.sm.AddSubscription(s.sub.Subscribe(event.TopicUsersMatchedCreated))
	s.sm.AddSubscription(s.sub.Subscribe(event.TopicGame))

	if err := s.init(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Service) init() error {

	// load games from cache
	games, err := s.cache.getGamesByServiceID(context.Background(), s.cfg.InstanceID)
	if err != nil {
		return err
	}

	for _, g := range games {
		if g.Status() == entity.GameStatusActive {
			if s.addGame(g) {
				sub := s.sub.Subscribe(event.TopicGame.SetResource(g.ID().String()))
				s.sm.AddSubscription(sub)
				s.l.Debug(fmt.Sprintf("subscribed to topic: '%s'", sub.Topic().String()))
			}

		}
	}

	return nil
}

func (gs *Service) addGame(g *entity.Game) bool {
	gs.gMu.Lock()
	defer gs.gMu.Unlock()
	if _, ok := gs.games[g.ID()]; ok {
		return false
	}
	gs.games[g.ID()] = g
	return true
}

func (gs *Service) getGame(id types.ObjectId) *entity.Game {
	gs.gMu.RLock()
	defer gs.gMu.RUnlock()
	return gs.games[id]
}

func (gs *Service) gameExists(id types.ObjectId) bool {
	gs.gMu.RLock()
	defer gs.gMu.RUnlock()
	_, ok := gs.games[id]
	return ok
}

func (gs *Service) removeGame(id types.ObjectId) {
	gs.gMu.Lock()
	defer gs.gMu.Unlock()
	delete(gs.games, id)
}

func (gs *Service) checkByPlayer(p types.ObjectId) bool {
	gs.gMu.RLock()
	defer gs.gMu.RUnlock()
	for _, g := range gs.games {
		if g.Player1().ID == p || g.Player2().ID == p {
			return true
		}
	}
	return false
}
