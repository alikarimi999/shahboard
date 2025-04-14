package game

import (
	"time"

	"github.com/alikarimi999/shahboard/event"
	"github.com/alikarimi999/shahboard/pkg/log"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	cfg Config

	sm    *event.SubscriptionManager
	gm    *gameManager
	ct    *playersConnectionTracker
	cache *redisGameCache

	live *liveGamesService

	pub event.Publisher
	sub event.Subscriber

	l log.Logger

	closeCh chan struct{}
}

func NewGameService(cfg Config, redis *redis.Client, pub event.Publisher, sub event.Subscriber,
	ws WsGateway, l log.Logger) (*Service, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	ct := newPlayersConnectionTracker(l, time.Duration(cfg.PlayerDisconnectTreshold)*time.Second)

	s := &Service{
		cfg:   cfg,
		ct:    ct,
		cache: newRedisGameCache(cfg.InstanceID, redis, 15*time.Minute, l),

		pub: pub,
		sub: sub,
		l:   l,

		closeCh: make(chan struct{}),
	}
	s.gm = newGameManager(s.cache, pub, ct, l)
	s.live = newLiveGamesService(s.cache, ws, l)

	s.sm = event.NewManager(l, s.handleEvents)
	s.sm.AddSubscription(s.sub.Subscribe(event.TopicUsersMatchedCreated))
	s.sm.AddSubscription(s.sub.Subscribe(event.TopicGame))

	// if err := s.init(); err != nil {
	// 	return nil, err
	// }

	return s, nil
}

// func (s *Service) init() error {

// 	// load games from cache
// 	games, err := s.cache.getGamesByServiceID(context.Background(), s.cfg.InstanceID)
// 	if err != nil {
// 		return err
// 	}

// 	for _, g := range games {
// 		if g.Status() == entity.GameStatusActive {
// 			if s.gm.addGame(g) {
// 				sub := s.sub.Subscribe(event.TopicGame.SetResource(g.ID().String()))
// 				s.sm.AddSubscription(sub)
// 				s.l.Debug(fmt.Sprintf("subscribed to topic: '%s'", sub.Topic().String()))
// 			}

// 		}
// 	}

// 	return nil
// }
