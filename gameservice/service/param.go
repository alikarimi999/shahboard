package game

import "github.com/alikarimi999/shahboard/types"

type GetGamePGNResponse struct {
	ID  types.ObjectId `json:"id"`
	PGN string         `json:"pgn"`
}

type GetGameFenResponse struct {
	ID  types.ObjectId `json:"id"`
	FEN string         `json:"fen"`
}
