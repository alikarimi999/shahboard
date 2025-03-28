package grpc

import (
	"context"

	pb "github.com/alikarimi999/shahboard/proto/rating/ratingpb"
	"github.com/alikarimi999/shahboard/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) GetUserRating(ctx context.Context, req *pb.GetUserRatingRequest) (*pb.GetUserRatingResponse, error) {
	userId, err := types.ParseObjectId(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user id")
	}

	rating, err := s.rating.GetUserRating(ctx, userId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user rating")
	}

	return &pb.GetUserRatingResponse{
		UserId:       rating.UserId.String(),
		CurrentScore: rating.CurrentScore,
		BestScore:    rating.BestScore,
		GamesPlayed:  rating.GamesPlayed,
		GamesWon:     rating.GamesWon,
		GamesLost:    rating.GamesLost,
		GamesDraw:    rating.GamesDraw,
		LastUpdated:  rating.LastUpdated.Unix(),
	}, nil

}
