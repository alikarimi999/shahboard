package game

import "github.com/alikarimi999/shahboard/types"

type GetLiveGamesDataResponse struct {
	List  []LiveGameDataResponse `json:"list"`
	Total int64                  `json:"total"`
}

type LiveGameDataResponse struct {
	GameID               types.ObjectId        `json:"game_id"`
	Player1              types.Player          `json:"player1"`
	Player2              types.Player          `json:"player2"`
	StartedAt            int64                 `json:"started_at"`
	ViewersNumber        int64                 `json:"viewers_number"`
	PriorityScore        int64                 `json:"priority_score"`
	PlayersDisconnection []PlayerDisconnection `json:"players_disconnection"`
}

type GetLiveGameIdByUserIdRequest struct {
	GameId types.ObjectId `json:"game_id"`
}

type GetGamePGNResponse struct {
	ID  types.ObjectId `json:"id"`
	PGN string         `json:"pgn"`

	PlayersDisconnection []PlayerDisconnection `json:"players_disconnection"`
}

type PlayerDisconnection struct {
	PlayerId       types.ObjectId `json:"player_id"`
	DisconnectedAt int64          `json:"disconnected_at"`
}

type GetGameFenResponse struct {
	ID  types.ObjectId `json:"id"`
	FEN string         `json:"fen"`
}
