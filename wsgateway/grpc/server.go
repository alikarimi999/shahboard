package grpc

import (
	"context"
	"fmt"
	"net"

	pb "github.com/alikarimi999/shahboard/proto/wsgateway/wsgatewaypb"
	"github.com/alikarimi999/shahboard/wsgateway/ws"

	"google.golang.org/grpc"
)

type Config struct {
	Port int `json:"port"`
}

type Server struct {
	cfg Config
	pb.UnimplementedWsGatewayServiceServer
	s        *grpc.Server
	wsServer *ws.Server
	lis      net.Listener
}

func NewServer(cfg Config, wsServer *ws.Server) (*Server, error) {
	if cfg.Port == 0 {
		return nil, fmt.Errorf("port is required")
	}
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		return nil, err
	}

	s := &Server{
		cfg:      cfg,
		s:        grpc.NewServer(),
		wsServer: wsServer,
		lis:      lis,
	}

	pb.RegisterWsGatewayServiceServer(s.s, s)

	return s, nil
}

func (s *Server) Run() error {
	return s.s.Serve(s.lis)
}

func (s *Server) GetLiveGamesViewersNumber(ctx context.Context,
	r *pb.LiveGamesViewersNumberRequest) (*pb.LiveGamesViewersNumberResponse, error) {

	m, err := s.wsServer.GetLiveGamesViewersNumber(ctx)
	if err != nil {
		return nil, err
	}

	mr := make(map[string]int64, len(m))
	for k, v := range m {
		mr[k.String()] = v
	}

	return &pb.LiveGamesViewersNumberResponse{
		GamesViewersNumber: mr,
	}, nil
}
