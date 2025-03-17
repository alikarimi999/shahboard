package game

import (
	"context"

	pb "github.com/alikarimi999/shahboard/proto/game/gamepb"
	"github.com/alikarimi999/shahboard/types"
	"google.golang.org/grpc"
)

type GameService struct {
	client pb.GameServiceClient
}

func NewService(client *grpc.ClientConn) *GameService {
	return &GameService{client: pb.NewGameServiceClient(client)}
}

// GetUserLiveGameID returns the live game ID for a given user ID.
// It returns types.ObjectZero if user doesn't have a live game.
func (s *GameService) GetUserLiveGameID(ctx context.Context, userId types.ObjectId) (types.ObjectId, error) {
	resp, err := s.client.GetUserLiveGameID(ctx, &pb.GetUserLiveGameIdRequest{UserId: userId.String()})
	if err != nil {
		return types.ObjectZero, err
	}

	if resp.GameId == "" {
		return types.ObjectZero, nil
	}

	return types.ParseObjectId(resp.GameId)
}
