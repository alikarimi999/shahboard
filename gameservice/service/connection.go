package game

import (
	"fmt"
	"maps"
	"sync"
	"time"

	"github.com/alikarimi999/shahboard/pkg/log"
	"github.com/alikarimi999/shahboard/types"
)

type playersConnectionTracker struct {
	mu   sync.RWMutex
	list map[types.ObjectId]map[types.ObjectId]time.Time // gameId -> playerId -> disconnectedAt

	disconnectThreshold time.Duration // in seconds

	l log.Logger
}

func newPlayersConnectionTracker(l log.Logger, disconnectThreshold time.Duration) *playersConnectionTracker {
	return &playersConnectionTracker{
		list:                make(map[types.ObjectId]map[types.ObjectId]time.Time),
		disconnectThreshold: disconnectThreshold,
		l:                   l,
	}
}

func (p *playersConnectionTracker) remove(gameId types.ObjectId, playerId types.ObjectId) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.list[gameId]; ok {
		delete(p.list[gameId], playerId)
		if len(p.list[gameId]) == 0 {
			delete(p.list, gameId)
		}
		p.l.Debug(fmt.Sprintf("player %s joined game %s", playerId, gameId))
	}
}

func (p *playersConnectionTracker) removeGame(gameId types.ObjectId) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.list, gameId)
}

func (p *playersConnectionTracker) get(gameId types.ObjectId) map[types.ObjectId]time.Time {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if _, ok := p.list[gameId]; ok {
		return p.list[gameId]
	}
	return nil
}

func (p *playersConnectionTracker) getAll() map[types.ObjectId]map[types.ObjectId]time.Time {
	p.mu.RLock()
	defer p.mu.RUnlock()

	list := make(map[types.ObjectId]map[types.ObjectId]time.Time)
	maps.Copy(list, p.list)

	return list
}

// getGamesNeedsToRemove returns the player who disconnected the earliest for each game. gameId -> playerId
func (p *playersConnectionTracker) getGamesNeedsToRemove() map[types.ObjectId]types.ObjectId {
	p.mu.Lock()
	defer p.mu.Unlock()

	t := time.Now().Add(-p.disconnectThreshold)
	endedGames := make(map[types.ObjectId]types.ObjectId)

	for gameId, players := range p.list {
		var earliestPlayer types.ObjectId
		var earliestTime time.Time

		for playerId, disconnectedAt := range players {
			if disconnectedAt.Before(t) {
				// If it's the first player found or the earliest disconnection
				if earliestTime.IsZero() || disconnectedAt.Before(earliestTime) {
					earliestPlayer = playerId
					earliestTime = disconnectedAt
				}
			}
		}

		// If we found a valid player to remove, store it
		if !earliestTime.IsZero() {
			endedGames[gameId] = earliestPlayer
		}
	}

	// Remove ended games from  list
	for gameId := range endedGames {
		delete(p.list, gameId)
	}

	return endedGames
}

func (p *playersConnectionTracker) add(gameId types.ObjectId, playerId types.ObjectId) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.list[gameId]; !ok {
		p.list[gameId] = make(map[types.ObjectId]time.Time)
	}
	p.list[gameId][playerId] = time.Now()
	p.l.Debug(fmt.Sprintf("player '%s' left game '%s'", playerId, gameId))
}
