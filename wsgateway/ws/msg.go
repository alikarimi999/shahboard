package ws

import (
	"github.com/alikarimi999/shahboard/types"
)

type msgType string

const (
	msgTypePlay msgType = "play"
	msgTypeView msgType = "view"
	msgTypeData msgType = "data"
	msgTypeErr  msgType = "err"
)

type msgBase struct {
	ID        types.ObjectId `json:"id"`
	Timestamp int64          `json:"timestamp"`
	Type      msgType        `json:"type"`
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
