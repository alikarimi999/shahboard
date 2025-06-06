package ws

import (
	"context"

	"github.com/alikarimi999/shahboard/types"
)

type GameService interface {
	// GetUserLiveGamePGN returns the live game PGN for a given user ID.
	// It returns nil if user doesn't have a live game.
	GetUserLiveGamePGN(ctx context.Context, userId types.ObjectId) (*GamePgn, error)

	// GetLiveGamePGN returns the live game PGN for a given game ID.
	// It returns nil if game doesn't exist.
	GetLiveGamePGN(ctx context.Context, gameId types.ObjectId) (*GamePgn, error)
}

type GamePgn struct {
	GameId                types.ObjectId
	Pgn                   string
	PlayersDisconnections []PlayerDisconnection
}

type PlayerDisconnection struct {
	PlayerId       types.ObjectId `json:"player_id"`
	DisconnectedAt int64          `json:"disconnected_at"`
}
