package game

import (
	"context"
	"fmt"

	pb "github.com/alikarimi999/shahboard/proto/game/gamepb"
	"github.com/alikarimi999/shahboard/types"
	"github.com/alikarimi999/shahboard/wsgateway/ws"
	"google.golang.org/grpc"
)

type Config struct {
	Target string `json:"target"`
}

type Service struct {
	client pb.GameServiceClient
}

func NewService(cfg Config, option grpc.DialOption) (*Service, error) {
	if cfg.Target == "" || option == nil {
		return nil, fmt.Errorf("address and dial option are required")
	}

	conn, err := grpc.NewClient(cfg.Target, option)
	if err != nil {
		return nil, err
	}

	return &Service{client: pb.NewGameServiceClient(conn)}, nil
}

// GetUserLiveGamePGN returns the live game PGN for a given user ID.
// It returns nil if user doesn't have a live game.
func (s *Service) GetUserLiveGamePGN(ctx context.Context, userId types.ObjectId) (*ws.GamePgn, error) {
	resp, err := s.client.GetUserLiveGamePGN(ctx, &pb.GetUserLiveGamePgnRequest{UserId: userId.String()})
	if err != nil {
		return nil, err
	}

	if resp.GameId == "" {
		return nil, nil
	}

	gameId, err := types.ParseObjectId(resp.GameId)
	if err != nil {
		return nil, nil
	}

	return &ws.GamePgn{
		GameId: gameId,
		Pgn:    resp.Pgn,
	}, nil
}
