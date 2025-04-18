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
)

type gameManager struct {
	gMu   sync.RWMutex
	games map[types.ObjectId]*entity.Game

	ct     *playersConnectionTracker
	pub    event.Publisher
	cache  *redisGameCache
	l      log.Logger
	stopCh chan struct{}
}

func newGameManager(cache *redisGameCache, pub event.Publisher, ct *playersConnectionTracker, l log.Logger) *gameManager {
	gm := &gameManager{
		games:  make(map[types.ObjectId]*entity.Game),
		ct:     ct,
		pub:    pub,
		cache:  cache,
		l:      l,
		stopCh: make(chan struct{}),
	}
	gm.run()

	return gm
}

func (gm *gameManager) run() {
	t := time.NewTicker(2 * time.Second)
	go func() {
		for {
			select {
			case <-gm.stopCh:
				t.Stop()
				return
			case <-t.C:
				gm.checkPlayersConnection()
			}
		}
	}()
}

// checkPlayersConnection find games that their players are disconnected are disconnected for more than disconnectThreshold
// and remove them from cache and publish game ended event.
func (gm *gameManager) checkPlayersConnection() {
	games := gm.getList()
	needToEnd := gm.ct.getGamesNeedsToRemove()

	endedGames := []*entity.Game{}
	endedGamesId := []types.ObjectId{}
	events := []event.Event{}

	// only proccess games that handling by this instance and let each instance handle its own games
	for _, game := range games {
		if playerId, ok := needToEnd[game.ID()]; ok {

			if game.Status() == entity.GameStatusDeactive || !game.PlayerLeft(playerId) {
				continue
			}

			gm.l.Debug(fmt.Sprintf("game %s ended, player %s left", game.ID(), playerId))

			endedGames = append(endedGames, game)

			events = append(events, event.EventGameEnded{
				GameID:    game.ID(),
				Player1:   game.Player1(),
				Player2:   game.Player2(),
				Outcome:   game.Outcome(),
				Desc:      entity.EndDescriptionPlayerLeft.String(),
				Timestamp: time.Now().Unix(),
			})

			endedGamesId = append(endedGamesId, game.ID())
		}
	}

	if len(endedGames) > 0 {
		if err := gm.cache.updateAndDeactivateGame(context.Background(), endedGames...); err != nil {
			gm.l.Error(err.Error())
		}
	}

	if len(events) > 0 {
		if err := gm.pub.Publish(events...); err != nil {
			gm.l.Error(err.Error())
		}
	}

	if len(endedGamesId) > 0 {
		gm.removeGame(endedGamesId...)
	}
}

func (gm *gameManager) addGame(g *entity.Game) bool {
	gm.gMu.Lock()
	defer gm.gMu.Unlock()
	if _, ok := gm.games[g.ID()]; ok {
		return false
	}
	gm.games[g.ID()] = g
	return true
}

func (gm *gameManager) getList() []*entity.Game {
	gm.gMu.RLock()
	defer gm.gMu.RUnlock()

	list := make([]*entity.Game, 0, len(gm.games))
	for _, game := range gm.games {
		list = append(list, game)
	}

	return list
}

func (gm *gameManager) getGame(id types.ObjectId) *entity.Game {
	gm.gMu.RLock()
	defer gm.gMu.RUnlock()
	return gm.games[id]
}

func (gm *gameManager) removeGame(ids ...types.ObjectId) {
	gm.gMu.Lock()
	defer gm.gMu.Unlock()

	for _, id := range ids {
		delete(gm.games, id)
	}
}

func (gm *gameManager) checkByPlayer(p types.ObjectId) bool {
	gm.gMu.RLock()
	defer gm.gMu.RUnlock()
	for _, g := range gm.games {
		if g.Player1().ID == p || g.Player2().ID == p {
			return true
		}
	}
	return false
}
