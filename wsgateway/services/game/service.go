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

	if resp == nil || resp.GameId == types.ObjectZero.String() || resp.GameId == "" {
		return nil, nil
	}

	gameId, err := types.ParseObjectId(resp.GameId)
	if err != nil {
		return nil, nil
	}

	ds := make([]ws.PlayerDisconnection, 0, len(resp.PlayersDisconnection))
	for _, v := range resp.PlayersDisconnection {
		ds = append(ds, ws.PlayerDisconnection{
			PlayerId:       types.ObjectId(v.PlayerId),
			DisconnectedAt: v.DisconnectedAt,
		})
	}

	return &ws.GamePgn{
		GameId:                gameId,
		Pgn:                   resp.Pgn,
		PlayersDisconnections: ds,
	}, nil
}

// GetLiveGamePGN returns the live game PGN for a given game ID.
// It returns nil if game doesn't exist.
func (s *Service) GetLiveGamePGN(ctx context.Context, gameId types.ObjectId) (*ws.GamePgn, error) {
	resp, err := s.client.GetLiveGamePGN(ctx, &pb.GetLiveGamePGNRequest{GameId: gameId.String()})
	if err != nil {
		return nil, err
	}

	if resp == nil || resp.GameId == types.ObjectZero.String() || resp.GameId == "" {
		return nil, nil
	}

	ds := make([]ws.PlayerDisconnection, 0, len(resp.PlayersDisconnection))
	for _, v := range resp.PlayersDisconnection {
		ds = append(ds, ws.PlayerDisconnection{
			PlayerId:       types.ObjectId(v.PlayerId),
			DisconnectedAt: v.DisconnectedAt,
		})
	}

	return &ws.GamePgn{
		GameId:                gameId,
		Pgn:                   resp.Pgn,
		PlayersDisconnections: ds,
	}, nil
}
