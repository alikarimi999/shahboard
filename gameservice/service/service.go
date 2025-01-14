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

	subManager *subscriptionManager

	gMu   sync.Mutex
	games map[types.ObjectId]*gameManager

	cache *redisGameCache

	p event.Publisher
	s event.Subscriber

	l log.Logger

	closeCh chan struct{}
	wg      sync.WaitGroup
}

func NewGameService(cfg Config, redis *redis.Client, p event.Publisher, s event.Subscriber, l log.Logger) (*Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	gs := &Service{
		cfg: cfg,

		games: make(map[types.ObjectId]*gameManager),
		cache: newRedisGameCache(cfg.InstanceID, redis, 15*time.Minute),
		p:     p,
		s:     s,
		l:     l,

		closeCh: make(chan struct{}),
	}

	gs.subManager = newSubscriptionManager(gs)

	if err := gs.init(); err != nil {
		return nil, err
	}

	gs.start()

	return gs, nil
}

func (gs *Service) start() {
	gs.wg.Add(1)
	go func() {
		defer gs.wg.Done()
		for range gs.closeCh {
			gs.subManager.stop()
			gs.gMu.Lock()
			for _, g := range gs.games {
				g.stop()
			}
			gs.gMu.Unlock()
		}
	}()
}

func (gs *Service) init() error {

	// subscribe to events
	gs.subscribeEvents(event.TopicMatch)

	// load games from cache
	games, err := gs.cache.getGamesByServiceID(context.Background(), gs.cfg.InstanceID)
	if err != nil {
		return err
	}

	for _, g := range games {
		if g.Status() == entity.GameStatusActive {
			gm := newGameManager(gs, g)
			gs.addGame(gm)

			// subscribe to the game
			topic := event.TopicGame.WithResource(gm.ID().String())
			gm.addSub(gs.s.Subscribe(topic))
			gs.l.Debug(fmt.Sprintf("subscribed to topic: '%s'", topic))

		}
	}

	return nil
}

func (gs *Service) addGame(g *gameManager) bool {
	gs.gMu.Lock()
	defer gs.gMu.Unlock()
	if _, ok := gs.games[g.ID()]; ok {
		return false
	}
	gs.games[g.ID()] = g
	return true
}

func (gs *Service) getGame(id types.ObjectId) *gameManager {
	gs.gMu.Lock()
	defer gs.gMu.Unlock()
	return gs.games[id]
}

func (gs *Service) removeGame(id types.ObjectId) {
	gs.gMu.Lock()
	defer gs.gMu.Unlock()
	delete(gs.games, id)
}

func (gs *Service) checkByPlayer(p types.ObjectId) bool {
	gs.gMu.Lock()
	defer gs.gMu.Unlock()
	for _, g := range gs.games {
		if g.Player1().ID == p || g.Player2().ID == p {
			return true
		}
	}
	return false
}
