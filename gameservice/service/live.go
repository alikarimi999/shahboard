package game

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

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
// separately building a list of 1000 live games by fetching the data
// from the Redis cache and ordering them based on players' scores
// and the number of viewers.
type liveGamesService struct {
	cache *redisGameCache
	ws    WsGateway

	mu   sync.RWMutex
	list []*LiveGameData

	l       log.Logger
	listCap int
	stopCh  chan struct{}
}

func newLiveGamesService(cache *redisGameCache, ws WsGateway, l log.Logger) *liveGamesService {
	ls := &liveGamesService{
		cache:   cache,
		ws:      ws,
		list:    make([]*LiveGameData, 0, 1000),
		listCap: 1000,
		l:       l,
		stopCh:  make(chan struct{}),
	}
	ls.run()

	return ls
}

func (s *liveGamesService) stop() {
	close(s.stopCh)
}
func (s *liveGamesService) run() {
	ticker := time.NewTicker(1 * time.Minute)
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

func (s *liveGamesService) updateLiveGames(ctx context.Context) error {

	games, err := s.cache.getLiveGamesData(ctx)
	if err != nil {
		return err
	}

	viwersNumber, err := s.ws.GetLiveGamesViewersNumber(ctx)
	if err != nil {
		return err
	}

	for _, g := range games {
		g.ViewersNumber = viwersNumber[g.GameID]
	}

	sortByPriorityScore(games)
	if len(games) > s.listCap {
		games = games[:s.listCap]
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.list = games

	return nil
}

func (s *liveGamesService) getLiveGames() []*LiveGameData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	newList := make([]*LiveGameData, len(s.list))
	copy(newList, s.list)
	return newList
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
	StartedAt     time.Time      `json:"started_at"`
	ViewersNumber int64          `json:"viewers_number"`
	PriorityScore int64          `json:"priority_score"`
}

func (g LiveGameData) encode() []byte {
	d, _ := json.Marshal(g)
	return d
}
