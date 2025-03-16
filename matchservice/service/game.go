package match

import (
	"context"

	"github.com/alikarimi999/shahboard/types"
)

type GameService interface {
	// GetUserLiveGameID returns the live game ID for a given user ID.
	// It returns types.ObjectZero if user doesn't have a live game.
	GetUserLiveGameID(ctx context.Context, userId types.ObjectId) (types.ObjectId, error)
}
