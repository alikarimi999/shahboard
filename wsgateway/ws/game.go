package ws

import (
	"context"

	"github.com/alikarimi999/shahboard/types"
)

type GameService interface {
	// GetUserLiveGamePGN returns the live game PGN for a given user ID.
	// It returns nil if user doesn't have a live game.
	GetUserLiveGamePGN(ctx context.Context, userId types.ObjectId) (*GamePgn, error)
}

type GamePgn struct {
	GameId types.ObjectId
	Pgn    string
}
