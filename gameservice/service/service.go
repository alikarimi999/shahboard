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
	games, err := gs.cache.GetGamesByServiceID(context.Background(), gs.cfg.InstanceID)
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

func (gs *Service) subscribeEvents(topics ...event.Topic) {
	for _, topic := range topics {
		gs.subManager.addSub(gs.s.Subscribe(topic))
		gs.l.Debug(fmt.Sprintf("subscribed to topic: '%s'", topic))
	}
}

// handleEventPlayersMatched handles the event of two players being matched
// and creates a new game if both players do not have any active games.
// This function is idempotent and safe to call concurrently from multiple
// instances of the GameService.
func (gs *Service) handleEventPlayersMatched(d *event.EventPlayersMatched) {

	// check if player is already in a game
	if gs.checkByPlayer(d.Player1) || gs.checkByPlayer(d.Player2) {
		gs.l.Debug("player is already in a game")
		return
	}

	// create a new game
	g := entity.NewGame(d.Player1, d.Player2, gs.cfg.DefaultGameSettings)

	// add the game to the cache
	if ok, err := gs.cache.AddGame(context.Background(), g); err != nil {
		gs.l.Error(err.Error())
	} else if !ok {
		gs.l.Debug(fmt.Sprintf("game already exists in cache: '%s'", g.ID()))
		return
	}
	gs.l.Debug(fmt.Sprintf("added game to cache: '%s'", g.ID()))

	gm := newGameManager(gs, g)
	gs.addGame(gm)

	// subscribe to the game
	topic := event.TopicGame.WithResource(gm.ID().String())
	gm.addSub(gs.s.Subscribe(topic))
	gs.l.Debug(fmt.Sprintf("subscribed to topic: '%s'", topic))

	// publish the game created event
	if err := gs.p.Publish(event.EventGameCreated{
		ID:        g.ID(),
		Player1:   g.Player1().ID,
		Player2:   g.Player2().ID,
		Timestamp: time.Now().Unix(),
	}); err != nil {
		gs.l.Error(err.Error())
	}
	gs.l.Debug(fmt.Sprintf("published game created event: '%s'", g.ID()))

	gs.l.Debug(fmt.Sprintf("game '%s' created by player '%s' as '%s' and player '%s' as '%s'",
		g.ID(), g.Player1().ID, g.Player1().Color, g.Player2().ID, g.Player2().Color))

	// TODO: add to repository concurrency control
}

func (gs *Service) handleEventGamePlayerMoved(d *event.EventGamePlayerMoved) {
	g := gs.getGame(d.ID)
	// if game is not manging by this instance, do nothing
	if g == nil {
		return
	}

	if g.Turn().ID != d.PlayerID {
		gs.l.Debug(fmt.Sprintf("it's not player '%s' turn", d.PlayerID))
		return
	}

	if err := g.Move(d.Move); err != nil {
		gs.l.Debug(fmt.Sprintf("player '%s' made an invalid move '%s' on game '%s'", d.PlayerID, d.Move, d.ID))
		return
	}

	if g.Outcome() != entity.NoOutcome {
		g.Deactivate()
		if err := gs.cache.UpdateAndDeactivateGame(context.Background(), g.Game); err != nil {
			gs.l.Error(err.Error())
			return
		}
		gs.p.Publish(event.EventGameMoveApproved{
			PlayerID:  d.PlayerID,
			ID:        d.ID,
			Move:      d.Move,
			Timestamp: time.Now().Unix(),
		},
			event.EventGameEnded{
				ID:        g.ID(),
				Player1:   g.Player1().ID,
				Player2:   g.Player2().ID,
				Timestamp: time.Now().Unix(),
			})

		gs.l.Debug(fmt.Sprintf("published game move approved event: '%s'", g.ID()))
		gs.l.Debug(fmt.Sprintf("published game ended event: '%s'", g.ID()))

	} else {
		if err := gs.cache.UpdateGameMove(context.Background(), g.Game); err != nil {
			gs.l.Debug(err.Error())
			return
		}

		gs.l.Debug(fmt.Sprintf("player '%s' made move '%s' on game '%s'", d.PlayerID, d.Move, g.ID()))

		gs.p.Publish(event.EventGameMoveApproved{
			PlayerID:  d.PlayerID,
			ID:        d.ID,
			Move:      d.Move,
			Timestamp: time.Now().Unix(),
		})
		gs.l.Debug(fmt.Sprintf("published game move approved event: '%s'", g.ID()))
	}

	// TODO: think about how to update the database
}

func (gs *Service) handleEventGamePlayerLeft(d *event.EventGamePlayerLeft) {
	g := gs.getGame(d.ID)
	if g == nil {
		return
	}

	g.Deactivate()
	if err := gs.cache.UpdateAndDeactivateGame(context.Background(), g.Game); err != nil {
		gs.l.Error(err.Error())
		return
	}
	gs.removeGame(d.ID)
	gs.l.Debug(fmt.Sprintf("removed game: '%s'", d.ID))

	gs.p.Publish(event.EventGameEnded{
		ID:        d.ID,
		Player1:   g.Player1().ID,
		Player2:   g.Player2().ID,
		Timestamp: time.Now().Unix(),
	})
	gs.l.Debug(fmt.Sprintf("published game ended event: '%s'", d.ID))

}

func (gs *Service) handleEvents(e event.Event) {
	switch e.GetTopic().Domain() {
	case event.DomainGame:
		switch e.GetAction() {
		case event.ActionGamePlayerMoved:
			gs.handleEventGamePlayerMoved(e.(*event.EventGamePlayerMoved))
		case event.ActionGamePlayerLeft:
			gs.handleEventGamePlayerLeft(e.(*event.EventGamePlayerLeft))
		}
	case event.DomainMatch:
		switch e.GetAction() {
		case event.ActionPlayersMatched:
			gs.handleEventPlayersMatched(e.(*event.EventPlayersMatched))
		}
	}
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

type Config struct {
	InstanceID          string              `json:"instance_id"`
	GamesCap            uint64              `json:"games_cap"`
	DefaultGameSettings entity.GameSettings `json:"default_game_settings"`
}

func (cfg Config) Validate() error {
	if cfg.InstanceID == "" {
		return fmt.Errorf("instance id is required")
	}

	if cfg.GamesCap == 0 {
		return fmt.Errorf("games cap is required")
	}

	if cfg.DefaultGameSettings.Time == 0 {
		return fmt.Errorf("default game setting time is required")
	}

	return nil
}
