package game

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"sync"
	"time"

	"github.com/alikarimi999/shahboard/event"
	"github.com/alikarimi999/shahboard/pkg/log"
	"github.com/alikarimi999/shahboard/types"
)

type WsGateway interface {
	GetLiveGamesViewersNumber(ctx context.Context) (map[types.ObjectId]int64, error)
}

// liveGamesService handles the management of live chess games,
// including providing a list of active games to clients.
// This service is intended to be separate from the game service
// and can be extended in the future to implement a recommendation system
// for game suggestions based on various filters and user preferences.
//
// Right now, each instance of the game service is responsible for
// separately building a list of 1000 live games by consuming game created event,
// and clean its cache by ended game event.
type liveGamesService struct {
	cache *redisGameCache
	ws    WsGateway

	mu         sync.RWMutex
	list       map[types.ObjectId]*LiveGameData // map by game ID
	sortedList []*LiveGameData

	l       log.Logger
	listCap int
	stopCh  chan struct{}
}

func newLiveGamesService(cache *redisGameCache, ws WsGateway, l log.Logger) *liveGamesService {
	ls := &liveGamesService{
		cache:   cache,
		ws:      ws,
		list:    make(map[types.ObjectId]*LiveGameData),
		listCap: 1000,
		l:       l,
		stopCh:  make(chan struct{}),
	}
	ls.run()

	return ls
}

// func (s *liveGamesService) stop() {
// 	close(s.stopCh)
// }

func (s *liveGamesService) add(g *LiveGameData) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.list[g.GameID] = g
	if len(s.sortedList) < s.listCap {
		s.sortedList = append(s.sortedList, g)
	}
}

func (s *liveGamesService) remove(gameID types.ObjectId) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.list, gameID)
	s.sortedList = slices.DeleteFunc(s.sortedList, func(g *LiveGameData) bool {
		return g.GameID == gameID
	})
}

func (s *liveGamesService) run() {
	ticker := time.NewTicker(30 * time.Second)
	go func() {

		for {
			select {
			case <-ticker.C:
				err := s.updateLiveGames(context.Background())
				if err != nil {
					s.l.Error(fmt.Sprintf("failed to update live games: %v", err))
				}
			case <-s.stopCh:
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *liveGamesService) getAll() []*LiveGameData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	newList := make([]*LiveGameData, 0, len(s.list))

	for _, g := range s.list {
		newList = append(newList, g)
	}
	return newList
}

func (s *liveGamesService) updateLiveGames(ctx context.Context) error {
	viwersNumber, err := s.ws.GetLiveGamesViewersNumber(ctx)
	if err != nil {
		return err
	}

	list := s.getAll()

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, g := range list {
		g.ViewersNumber = viwersNumber[g.GameID]
	}

	sortByPriorityScore(list)
	if len(list) > s.listCap {
		list = list[:s.listCap]
	}

	s.sortedList = list

	return nil
}

func (s *liveGamesService) getLiveGamesSorted() (list []*LiveGameData, total int64) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	newList := make([]*LiveGameData, len(s.sortedList))
	copy(newList, s.sortedList)
	return newList, int64(len(s.list))
}

func sortByPriorityScore(list []*LiveGameData) {
	sort.Slice(list, func(i, j int) bool {
		list[i].PriorityScore = calcPriorityScore(list[i])
		list[j].PriorityScore = calcPriorityScore(list[j])
		return list[i].PriorityScore > list[j].PriorityScore
	})
}

func calcPriorityScore(g *LiveGameData) int64 {
	return g.Player1.Score + g.Player2.Score + g.ViewersNumber
}

type LiveGameData struct {
	GameID        types.ObjectId `json:"game_id"`
	Player1       types.Player   `json:"player1"`
	Player2       types.Player   `json:"player2"`
	StartedAt     int64          `json:"started_at"`
	ViewersNumber int64          `json:"viewers_number"`
	PriorityScore int64          `json:"priority_score"`
}

func (gs *Service) handleEventGameCreated(e *event.EventGameCreated) {
	gs.live.add(&LiveGameData{
		GameID:    e.GameID,
		Player1:   e.Player1,
		Player2:   e.Player2,
		StartedAt: e.Timestamp,
	})
}

func (gs *Service) handleEventGameEnded(e *event.EventGameEnded) {
	gs.live.remove(e.GameID)
}
