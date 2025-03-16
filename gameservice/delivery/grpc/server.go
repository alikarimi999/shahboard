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

	if res.IsZero() {
		return nil, status.Errorf(codes.NotFound, "user live game not found")
	}

	return &pb.GetUserLiveGameIdResponse{
		GameId: res.String(),
	}, nil
}

func (s *Server) GetUserLiveGamePGN(ctx context.Context, req *pb.GetUserLiveGamePgnRequest) (*pb.GetUserLiveGamePgnResponse, error) {
	userId, err := types.ParseObjectId(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user id: %v", err)
	}

	res, err := s.svc.GetLiveGamePgnByUserID(ctx, userId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user live game: %v", err)
	}

	if res.ID.IsZero() {
		return nil, status.Errorf(codes.NotFound, "user live game not found")
	}

	return &pb.GetUserLiveGamePgnResponse{
		GameId: res.ID.String(),
		Pgn:    res.PGN,
	}, nil
}
