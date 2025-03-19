package game

import (
	"context"

	"github.com/alikarimi999/shahboard/gameservice/entity"
	"github.com/alikarimi999/shahboard/types"
)

func (s *Service) GetLiveGames(ctx context.Context) ([]types.ObjectId, error) {
	return s.cache.getGamesIDs(ctx)
}

func (s *Service) GetLiveGameIdByUserId(ctx context.Context, id types.ObjectId) (types.ObjectId, error) {
	return s.cache.getGameIdByUserID(ctx, id)

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

	game, err := s.cache.getGameByID(ctx, gameId)
	if err != nil {
		return GetGamePGNResponse{}, err
	}

	if game == nil || game.Status() == entity.GameStatusDeactive {
		return GetGamePGNResponse{ID: types.ObjectZero}, nil
	}

	return GetGamePGNResponse{ID: game.ID(), PGN: game.PGN()}, nil
}

// GetLiveGamePGN returns the game ID and PGN for a given game ID.
func (s *Service) GetLiveGamePGN(ctx context.Context, id types.ObjectId) (GetGamePGNResponse, error) {
	game, err := s.cache.getGameByID(ctx, id)
	if err != nil {
		return GetGamePGNResponse{}, err
	}

	if game == nil || game.Status() == entity.GameStatusDeactive {
		return GetGamePGNResponse{ID: types.ObjectZero}, nil
	}

	return GetGamePGNResponse{ID: game.ID(), PGN: game.PGN()}, nil
}

func (s *Service) GetLiveGameByID(ctx context.Context, id types.ObjectId) (GetGamePGNResponse, error) {
	game, err := s.cache.getGameByID(ctx, id)
	if err != nil {
		return GetGamePGNResponse{}, err
	}

	if game == nil || game.Status() == entity.GameStatusDeactive {
		return GetGamePGNResponse{ID: types.ObjectZero}, nil
	}

	return GetGamePGNResponse{ID: game.ID(), PGN: game.PGN()}, nil
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

func (s *Service) GetGamePGN(ctx context.Context, id types.ObjectId) (GetGamePGNResponse, error) {
	g, err := s.getGameByID(ctx, id)
	if err != nil {
		return GetGamePGNResponse{}, err
	}
	return GetGamePGNResponse{ID: g.ID(), PGN: g.PGN()}, nil
}

func (s *Service) getGameByID(ctx context.Context, id types.ObjectId) (*entity.Game, error) {
	if g := s.getGame(id); g != nil {
		return g, nil
	}

	return s.cache.getGameByID(ctx, id)
}
