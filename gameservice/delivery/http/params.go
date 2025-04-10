package http

import (
	game "github.com/alikarimi999/shahboard/gameservice/service"
	"github.com/alikarimi999/shahboard/types"
)

type list struct {
	List []interface{} `json:"list"`
}

type GetLiveGameDataResponse struct {
	List  []*game.LiveGameData `json:"list"`
	Total int64                `json:"total"`
}

type GetLiveGameIdByUserIdRequest struct {
	GameId types.ObjectId `json:"game_id"`
}
