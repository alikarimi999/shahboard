package ws

import (
	"github.com/alikarimi999/shahboard/types"
)

type MsgData interface {
	Type() MsgType
	Encode() []byte
}

type MsgType string

const (
	MsgTypeWelcome      MsgType = "welcome"
	MsgTypeFindMatch    MsgType = "find_match"
	MsgTypeView         MsgType = "view"
	MsgTypeData         MsgType = "data"
	MsgTypeError        MsgType = "err"
	MsgTypePing         MsgType = "ping"
	MsgTypePong         MsgType = "pong"
	MsgTypeGameCreate   MsgType = "game_create"
	MsgTypePlayerMove   MsgType = "player_move"
	MsgTypeMoveApproved MsgType = "move_approved"
	MsgTypeGameEnd      MsgType = "game_end"
	MsgTypeMove         MsgType = "move"
)

type MsgBase struct {
	ID        types.ObjectId `json:"id"`
	Type      MsgType        `json:"type"`
	Timestamp int64          `json:"timestamp"`
}

type ClientMsg struct {
	MsgBase
	Data MsgData `json:"data"`
}

type ServerMsg struct {
	MsgBase
	Data []byte `json:"data"`
}
