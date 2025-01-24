package ws

import (
	"github.com/alikarimi999/shahboard/types"
)

type msgType string

const (
	msgTypeWelcome msgType = "welcome"
	msgTypePlay    msgType = "play"
	msgTypeView    msgType = "view"
	msgTypeData    msgType = "data"
	msgTypeErr     msgType = "err"
	msgTypePing    msgType = "ping"
	msgTypePong    msgType = "pong"
)

type msgBase struct {
	ID        types.ObjectId `json:"id"`
	Type      msgType        `json:"type"`
	Timestamp int64          `json:"timestamp"`
}

type clientMsg struct {
	msgBase
	Data []byte `json:"data"`
}

type serverMsg struct {
	msgBase
	Data []byte `json:"data"`
}

type serverMsgData interface {
	Encode() []byte
}

type serverMsgErr string

func (e serverMsgErr) Encode() []byte {
	return []byte(e)
}

type playCmdData struct {
	GameId types.ObjectId `json:"game_id"`
}

type viewCmdData struct {
	GameId types.ObjectId `json:"game_id"`
}
