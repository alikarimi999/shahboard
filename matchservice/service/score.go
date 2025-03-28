package match

import "github.com/alikarimi999/shahboard/types"

type RatingService interface {
	GetUserLevel(id types.ObjectId) (types.Level, error)
}
