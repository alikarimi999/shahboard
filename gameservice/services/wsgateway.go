package services

import (
	"context"

	pb "github.com/alikarimi999/shahboard/proto/wsgateway/wsgatewaypb"
	"github.com/alikarimi999/shahboard/types"
	"google.golang.org/grpc"
)

type WsGatewayService struct {
	client pb.WsGatewayServiceClient
}

func NewService(client *grpc.ClientConn) *WsGatewayService {
	return &WsGatewayService{client: pb.NewWsGatewayServiceClient(client)}
}

func (s *WsGatewayService) GetLiveGamesViewersNumber(ctx context.Context) (map[types.ObjectId]int64, error) {
	resp, err := s.client.GetLiveGamesViewersNumber(ctx, &pb.LiveGamesViewersNumberRequest{})
	if err != nil {
		return nil, err
	}

	res := make(map[types.ObjectId]int64, len(resp.GamesViewersNumber))
	for k, v := range resp.GamesViewersNumber {
		id, err := types.ParseObjectId(k)
		if err != nil {
			return nil, err
		}
		res[id] = v
	}

	return res, nil
}
