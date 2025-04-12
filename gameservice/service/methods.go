package game

import (
	"context"
	"fmt"
	"time"

	"github.com/alikarimi999/shahboard/event"
	"github.com/alikarimi999/shahboard/gameservice/entity"
	"github.com/alikarimi999/shahboard/types"
)

func (s *Service) GetLiveGamesIDs(ctx context.Context) ([]types.ObjectId, error) {
	return s.cache.getGamesIDs(ctx)
}

func (s *Service) GetLiveGamesData(ctx context.Context) (GetLiveGamesDataResponse, error) {
	list, total := s.live.getLiveGamesSorted()
	connections := s.ct.getAll()
	res := GetLiveGamesDataResponse{
		List:  make([]LiveGameDataResponse, 0, len(list)),
		Total: total,
	}

	for _, g := range list {
		l := LiveGameDataResponse{
			GameID:        g.GameID,
			Player1:       g.Player1,
			Player2:       g.Player2,
			StartedAt:     g.StartedAt,
			ViewersNumber: g.ViewersNumber,
			PriorityScore: g.PriorityScore,
		}
		if conn, ok := connections[g.GameID]; ok {
			for p, t := range conn {
				l.PlayersDisconnection = append(l.PlayersDisconnection, PlayerDisconnection{
					PlayerId:       p,
					DisconnectedAt: t.Unix(),
				})
			}
		}
		res.List = append(res.List, l)
	}

	return res, nil
}

func (s *Service) GetLiveGameIdByUserId(ctx context.Context, id types.ObjectId) (GetLiveGameIdByUserIdRequest, error) {
	gameId, err := s.cache.getGameIdByUserID(ctx, id)
	if err != nil {
		return GetLiveGameIdByUserIdRequest{}, err
	}

	return GetLiveGameIdByUserIdRequest{GameId: gameId}, nil
}

// currently use event for proccessing player resign request.
//
// TODO: make this ready for a multi instance setup
func (s *Service) ResingByPlayer(ctx context.Context, gameId, playerId types.ObjectId) error {
	game, err := s.getGameByID(ctx, gameId)
	if err != nil {
		return err
	}

	if game == nil || game.Status() == entity.GameStatusDeactive {
		return fmt.Errorf("game not found")
	}

	if ok := game.Resign(playerId); !ok {
		return fmt.Errorf("player is not in the game")
	}

	s.l.Debug(fmt.Sprintf("player '%s' resigned game '%s'", playerId, gameId))

	if err := s.cache.updateAndDeactivateGame(context.Background(), game); err != nil {
		s.l.Error(err.Error())
		return err
	}

	s.gm.removeGame(gameId)

	if err := s.pub.Publish(event.EventGameEnded{
		ID:        types.NewObjectId(),
		GameID:    gameId,
		Player1:   game.Player1(),
		Player2:   game.Player2(),
		Outcome:   game.Outcome(),
		Desc:      entity.EndDescriptionPlayerResigned.String(),
		Timestamp: time.Now().Unix(),
	}); err != nil {
		s.l.Error(err.Error())
	}

	s.l.Debug(fmt.Sprintf("published game ended event: '%s'", gameId))

	return nil
}

// GetLiveGamePgnByUserID returns the game ID and PGN for a given user ID.
// It first checks the cache for the game ID associated with the user ID.
// If the game ID is found, it retrieves the game from the cache.
// If the game is not found, it returns an empty response with a zero ID.
func (s *Service) GetLiveGamePgnByUserID(ctx context.Context, id types.ObjectId) (GetGamePGNResponse, error) {
	gameId, err := s.cache.getGameIdByUserID(ctx, id)
	if err != nil {
		return GetGamePGNResponse{}, err
	}

	if gameId.IsZero() {
		return GetGamePGNResponse{ID: types.ObjectZero}, nil
	}

	return s.getGamePGN(ctx, gameId)
}

// GetLiveGamePGN returns the game ID and PGN for a given game ID.
func (s *Service) GetLiveGamePGN(ctx context.Context, id types.ObjectId) (GetGamePGNResponse, error) {
	return s.getGamePGN(ctx, id)
}

func (s *Service) getGamePGN(ctx context.Context, id types.ObjectId) (GetGamePGNResponse, error) {
	game, err := s.cache.getGameByID(ctx, id)
	if err != nil {
		return GetGamePGNResponse{}, err
	}

	if game == nil || game.Status() == entity.GameStatusDeactive {
		return GetGamePGNResponse{ID: types.ObjectZero}, nil
	}

	res := GetGamePGNResponse{ID: game.ID(), PGN: game.PGN()}
	for p, t := range s.ct.get(game.ID()) {
		res.PlayersDisconnection = append(res.PlayersDisconnection, PlayerDisconnection{
			PlayerId:       p,
			DisconnectedAt: t.Unix(),
		})
	}

	return res, nil
}

func (s *Service) GetGamesFEN(ctx context.Context, ids []types.ObjectId) (map[types.ObjectId]string, error) {
	games, err := s.cache.getGamesByID(ctx, ids)
	if err != nil {
		return nil, err
	}

	res := make(map[types.ObjectId]string)
	for _, g := range games {
		res[g.ID()] = g.FEN()
	}

	return res, nil
}

func (s *Service) getGameByID(ctx context.Context, id types.ObjectId) (*entity.Game, error) {
	if g := s.gm.getGame(id); g != nil {
		return g, nil
	}

	return s.cache.getGameByID(ctx, id)
}
