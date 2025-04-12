package grpc

import (
	"context"
	"fmt"
	"net"

	game "github.com/alikarimi999/shahboard/gameservice/service"
	pb "github.com/alikarimi999/shahboard/proto/game/gamepb"
	"github.com/alikarimi999/shahboard/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
)

type Config struct {
	Port int `json:"port"`
}

type Server struct {
	cfg Config
	pb.UnimplementedGameServiceServer
	svc *game.Service
	s   *grpc.Server
	lis net.Listener
}

func NewServer(cfg Config, svc *game.Service) (*Server, error) {
	if cfg.Port == 0 {
		return nil, fmt.Errorf("port is required")
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		return nil, err
	}

	s := &Server{
		cfg: cfg,
		svc: svc,
		s:   grpc.NewServer(),
		lis: lis,
	}

	pb.RegisterGameServiceServer(s.s, s)
	return s, nil
}

func (s *Server) Run() error {
	return s.s.Serve(s.lis)
}

func (s *Server) GetUserLiveGameID(ctx context.Context, req *pb.GetUserLiveGameIdRequest) (*pb.GetUserLiveGameIdResponse, error) {
	userId, err := types.ParseObjectId(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user id: %v", err)
	}
	res, err := s.svc.GetLiveGameIdByUserId(ctx, userId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user live game: %v", err)
	}

	if res.GameId.IsZero() {
		return &pb.GetUserLiveGameIdResponse{
			GameId: types.ObjectZero.String(),
		}, nil
	}

	return &pb.GetUserLiveGameIdResponse{
		GameId: res.GameId.String(),
	}, nil
}

func (s *Server) GetUserLiveGamePGN(ctx context.Context, req *pb.GetUserLiveGamePgnRequest) (*pb.GetLiveGamePGNResponse, error) {
	userId, err := types.ParseObjectId(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user id: %v", err)
	}

	res, err := s.svc.GetLiveGamePgnByUserID(ctx, userId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user live game: %v", err)
	}

	if res.ID.IsZero() {
		return &pb.GetLiveGamePGNResponse{
			GameId: types.ObjectZero.String(),
		}, nil
	}

	ds := make([]*pb.PlayerDisconnection, 0, len(res.PlayersDisconnection))
	for _, v := range res.PlayersDisconnection {
		ds = append(ds, &pb.PlayerDisconnection{
			PlayerId:       v.PlayerId.String(),
			DisconnectedAt: v.DisconnectedAt,
		})
	}

	return &pb.GetLiveGamePGNResponse{
		GameId:               res.ID.String(),
		Pgn:                  res.PGN,
		PlayersDisconnection: ds,
	}, nil
}

func (s *Server) GetLiveGamePGN(ctx context.Context, req *pb.GetLiveGamePGNRequest) (*pb.GetLiveGamePGNResponse, error) {
	gameId, err := types.ParseObjectId(req.GameId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid game id: %v", err)
	}

	res, err := s.svc.GetLiveGamePGN(ctx, gameId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get live game: %v", err)
	}

	if res.ID.IsZero() {
		return &pb.GetLiveGamePGNResponse{
			GameId: types.ObjectZero.String(),
		}, nil
	}

	ds := make([]*pb.PlayerDisconnection, 0, len(res.PlayersDisconnection))
	for _, v := range res.PlayersDisconnection {
		ds = append(ds, &pb.PlayerDisconnection{
			PlayerId:       v.PlayerId.String(),
			DisconnectedAt: v.DisconnectedAt,
		})
	}

	return &pb.GetLiveGamePGNResponse{
		GameId:               res.ID.String(),
		Pgn:                  res.PGN,
		PlayersDisconnection: ds,
	}, nil
}
