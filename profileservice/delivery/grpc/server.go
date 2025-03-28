package grpc

import (
	"fmt"
	"net"

	"github.com/alikarimi999/shahboard/profileservice/service/rating"
	pb "github.com/alikarimi999/shahboard/proto/rating/ratingpb"

	"google.golang.org/grpc"
)

type Config struct {
	Port int `json:"port"`
}

type Server struct {
	cfg Config
	pb.UnimplementedRatingServiceServer
	rating *rating.Service
	s      *grpc.Server
	lis    net.Listener
}

func NewServer(cfg Config, svc *rating.Service) (*Server, error) {
	if cfg.Port == 0 {
		return nil, fmt.Errorf("port is required")
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		return nil, err

	}
	s := &Server{
		cfg:    cfg,
		rating: svc,
		s:      grpc.NewServer(),
		lis:    lis,
	}

	pb.RegisterRatingServiceServer(s.s, s)
	return s, nil
}

func (s *Server) Run() error {
	return s.s.Serve(s.lis)
}
