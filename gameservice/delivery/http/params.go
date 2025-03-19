package http

import "github.com/alikarimi999/shahboard/types"

type list struct {
	List []interface{} `json:"list"`
}

type GetLiveGameIdByUserIdRequest struct {
	GameId types.ObjectId `json:"game_id"`
}
