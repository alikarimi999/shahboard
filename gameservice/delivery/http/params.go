package http

import "github.com/alikarimi999/shahboard/types"

type getGamesFENRequest struct {
	Games []types.ObjectId `json:"games"`
}

type fen struct {
	ID  types.ObjectId `json:"id"`
	FEN string         `json:"fen"`
}

type getGamePGNResponse struct {
	ID  types.ObjectId `json:"id"`
	PGN string         `json:"pgn"`
}
type list struct {
	List []interface{} `json:"list"`
}
