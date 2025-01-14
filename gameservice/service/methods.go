package game

import (
	"context"

	"github.com/alikarimi999/shahboard/gameservice/entity"
	"github.com/alikarimi999/shahboard/types"
)

func (s *Service) GetLiveGames(ctx context.Context) ([]types.ObjectId, error) {
	return s.cache.getGamesIDs(ctx)
}

func (s *Service) GetGamesByID(ctx context.Context, ids []types.ObjectId) ([]*entity.Game, error) {
	return s.cache.getGamesByID(ctx, ids)
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

func (s *Service) GetGamePGN(ctx context.Context, id types.ObjectId) (string, error) {
	g, err := s.getGameByID(ctx, id)
	if err != nil {
		return "", err
	}
	return g.PGN(), nil
}

func (s *Service) getGameByID(ctx context.Context, id types.ObjectId) (*entity.Game, error) {
	if g := s.getGame(id); g != nil {
		return g.Game, nil
	}

	return s.cache.getGameByID(ctx, id)
}
