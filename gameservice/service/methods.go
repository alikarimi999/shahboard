package game

import (
	"context"
	"fmt"

	"github.com/alikarimi999/shahboard/gameservice/entity"
	"github.com/alikarimi999/shahboard/types"
)

func (s *Service) GetLiveGames(ctx context.Context) ([]types.ObjectId, error) {
	return s.cache.getGamesIDs(ctx)
}

func (s *Service) GetLiveGameByUserID(ctx context.Context, id types.ObjectId) (GetGamePGNResponse, error) {
	gameId, err := s.cache.getGameByUserID(ctx, id)
	if err != nil {
		return GetGamePGNResponse{}, err
	}

	if gameId.IsZero() {
		return GetGamePGNResponse{}, fmt.Errorf("game not found")
	}

	game, err := s.cache.getGameByID(ctx, gameId)
	if err != nil {
		return GetGamePGNResponse{}, err
	}

	return GetGamePGNResponse{ID: game.ID(), PGN: game.PGN()}, nil
}

func (s *Service) GetLiveGameByID(ctx context.Context, id types.ObjectId) (GetGamePGNResponse, error) {
	game, err := s.cache.getGameByID(ctx, id)
	if err != nil {
		return GetGamePGNResponse{}, err
	}

	if game == nil {
		return GetGamePGNResponse{}, fmt.Errorf("game not found")
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
		return g.Game, nil
	}

	return s.cache.getGameByID(ctx, id)
}
