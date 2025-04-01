package match

import "github.com/alikarimi999/shahboard/types"

type RatingService interface {
	GetUserScore(id types.ObjectId) (int64, error)
}
