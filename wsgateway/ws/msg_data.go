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

type DataGameChatCreated struct {
	event.EventGameChatCreated
}

func (d DataGameChatCreated) Type() MsgType {
	return MsgTypeChatCreated
}

func (d DataGameChatCreated) Encode() []byte {
	b, _ := json.Marshal(d)
	return b
}

type DataGameChatMsgSend struct {
	event.EventGameChatMsgeSent
}

func (d DataGameChatMsgSend) Type() MsgType {
	return MsgTypeChatMsgSend
}

func (d DataGameChatMsgSend) Encode() []byte {
	b, _ := json.Marshal(d)
	return b
}

type DataGameChatMsgApproved struct {
	event.EventGameChatMsgApproved
}

func (d DataGameChatMsgApproved) Type() MsgType {
	return MsgTypeChatMsgApproved
}

func (d DataGameChatMsgApproved) Encode() []byte {
	b, _ := json.Marshal(d)
	return b
}
