package ws

import (
	"encoding/json"

	"github.com/alikarimi999/shahboard/event"
	"github.com/alikarimi999/shahboard/types"
)

type DataFindMatchRequest struct {
	event.EventUsersMatchCreated
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
	return MsgTypeViewGame
}

func (d DataGameViewRequest) Encode() []byte {
	b, _ := json.Marshal(d)
	return b
}

type DataGamePlayerMoveRequest struct {
	event.EventGamePlayerMoved
}

func (d DataGamePlayerMoveRequest) Type() MsgType {
	return MsgTypePlayerMove
}

func (d DataGamePlayerMoveRequest) Encode() []byte {
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

type DataResumeGameRequest struct {
	GameId types.ObjectId `json:"game_id"`
}

func (m DataResumeGameRequest) Type() MsgType {
	return MsgTypeResumeGame
}

func (m DataResumeGameRequest) Encode() []byte {
	b, _ := json.Marshal(m)
	return b
}

type DataResumeGameResponse struct {
	GameId types.ObjectId `json:"game_id"`
	Pgn    string         `json:"pgn"`
}

func (m DataResumeGameResponse) Type() MsgType {
	return MsgTypeResumeGame
}

func (m DataResumeGameResponse) Encode() []byte {
	b, _ := json.Marshal(m)
	return b
}

type DataGameViewResponse struct {
	GameId types.ObjectId `json:"game_id"`
	Pgn    string         `json:"pgn"`
}

func (m DataGameViewResponse) Type() MsgType {
	return MsgTypeViewGame
}
func (m DataGameViewResponse) Encode() []byte {
	b, _ := json.Marshal(m)
	return b
}

type DataViwersListResponse struct {
	GameId types.ObjectId   `json:"game_id"`
	List   []types.ObjectId `json:"list"`
}

func (m DataViwersListResponse) Type() MsgType {
	return MsgTypeViewersList
}

func (m DataViwersListResponse) Encode() []byte {
	b, _ := json.Marshal(m)
	return b
}

type DataGamePlayerResignRequest struct {
	event.EventGamePlayerResigned
}

func (m DataGamePlayerResignRequest) Type() MsgType {
	return MsgTypePlayerResigned
}

func (m DataGamePlayerResignRequest) Encode() []byte {
	b, _ := json.Marshal(m)
	return b
}
