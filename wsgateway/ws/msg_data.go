package ws

import (
	"encoding/json"

	"github.com/alikarimi999/shahboard/event"
	"github.com/alikarimi999/shahboard/types"
)

type dataFindMatchRequest struct {
	event.EventUsersMatchCreated
}

func (d dataFindMatchRequest) Type() MsgType {
	return MsgTypeFindMatch
}

func (d dataFindMatchRequest) Encode() []byte {
	b, _ := json.Marshal(d)
	return b
}

type dataGameViewRequest struct {
	GameId types.ObjectId `json:"game_id"`
}

func (d dataGameViewRequest) Type() MsgType {
	return MsgTypeView
}

func (d dataGameViewRequest) Encode() []byte {
	b, _ := json.Marshal(d)
	return b
}

type dataGamePlayerMoveRequest struct {
	ID        types.ObjectId `json:"id"`
	GameID    types.ObjectId `json:"game_id"`
	PlayerID  types.ObjectId `json:"player_id"`
	Move      string         `json:"move"`
	Timestamp int64          `json:"timestamp"`
}

func (d dataGamePlayerMoveRequest) Type() MsgType {
	return MsgTypePlayerMove
}

func (d dataGamePlayerMoveRequest) Encode() []byte {
	b, _ := json.Marshal(d)
	return b
}

type dataGameChatMsgSend struct {
	event.EventGameChatMsgeSent
}

func (d dataGameChatMsgSend) Type() MsgType {
	return MsgTypeChatMsgSend
}

func (d dataGameChatMsgSend) Encode() []byte {
	b, _ := json.Marshal(d)
	return b
}

type dataResumeGameRequest struct {
	GameId types.ObjectId `json:"game_id"`
}

func (m dataResumeGameRequest) Type() MsgType {
	return MsgTypeResumeGame
}

func (m dataResumeGameRequest) Encode() []byte {
	b, _ := json.Marshal(m)
	return b
}
