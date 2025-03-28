package services

import (
	"context"

	"github.com/alikarimi999/shahboard/pkg/elo"
	pb "github.com/alikarimi999/shahboard/proto/rating/ratingpb"
	"github.com/alikarimi999/shahboard/types"
	"google.golang.org/grpc"
)

type RatingService struct {
	c pb.RatingServiceClient
}

func NewRatingService(client *grpc.ClientConn) *RatingService {
	return &RatingService{
		c: pb.NewRatingServiceClient(client),
	}
}

func (s *RatingService) GetUserLevel(id types.ObjectId) (types.Level, error) {
	res, err := s.c.GetUserRating(context.Background(), &pb.GetUserRatingRequest{UserId: id.String()})
	if err != nil {
		return 0, err
	}

	return elo.GetPlayerLevel(res.CurrentScore), nil
}
