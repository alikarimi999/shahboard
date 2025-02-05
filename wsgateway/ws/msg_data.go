package ws

import (
	"encoding/json"

	"github.com/alikarimi999/shahboard/event"
	"github.com/alikarimi999/shahboard/types"
)

type DataFindMatchRequest struct {
	event.EventUsersMatched
}

func (d DataFindMatchRequest) Type() MsgType {
	return MsgTypeFindMatch
}

func (d DataFindMatchRequest) Encode() []byte {
	b, _ := json.Marshal(d)
	return b
}

type DataGameViewRequest struct {
	GameId types.ObjectId `json:"game_id"`
}

func (d DataGameViewRequest) Type() MsgType {
	return MsgTypeView
}

func (d DataGameViewRequest) Encode() []byte {
	b, _ := json.Marshal(d)
	return b
}

type DataGamePlayerMoveRequest struct {
	ID        types.ObjectId `json:"id"`
	GameID    types.ObjectId `json:"game_id"`
	PlayerID  types.ObjectId `json:"player_id"`
	Move      string         `json:"move"`
	Timestamp int64          `json:"timestamp"`
}

func (d DataGamePlayerMoveRequest) Type() MsgType {
	return MsgTypePlayerMove
}

func (d DataGamePlayerMoveRequest) Encode() []byte {
	b, _ := json.Marshal(d)
	return b
}
