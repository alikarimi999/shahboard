package game

import (
	"context"

	pb "github.com/alikarimi999/shahboard/proto/game/gamepb"
	"github.com/alikarimi999/shahboard/types"
	"github.com/alikarimi999/shahboard/wsgateway/ws"
	"google.golang.org/grpc"
)

type Service struct {
	client pb.GameServiceClient
}

func NewService(client *grpc.ClientConn) *Service {
	return &Service{client: pb.NewGameServiceClient(client)}
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
