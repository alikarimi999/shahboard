package ws

import (
	"github.com/alikarimi999/shahboard/types"
)

type MsgType string

const (
	MsgTypeWelcome      MsgType = "welcome"
	MsgTypeFindMatch    MsgType = "find_match"
	MsgTypeView         MsgType = "view"
	MsgTypeData         MsgType = "data"
	MsgTypeError        MsgType = "err"
	MsgTypeGameCreate   MsgType = "game_created"
	MsgTypePlayerMove   MsgType = "player_moved"
	MsgTypeMoveApproved MsgType = "move_approved"
	MsgTypeGameEnd      MsgType = "game_ended"
)

type MsgBase struct {
	ID        types.ObjectId `json:"id"`
	Type      MsgType        `json:"type"`
	Timestamp int64          `json:"timestamp"`
}

type Msg struct {
	MsgBase
	Data []byte `json:"data"`
}
