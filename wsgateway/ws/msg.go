package ws

import (
	"encoding/json"

	"github.com/alikarimi999/shahboard/types"
)

type MsgType string

const (
	MsgTypeWelcome                 MsgType = "welcome"
	MsgTypeFindMatch               MsgType = "find_match"
	MsgTypeView                    MsgType = "view"
	MsgTypeData                    MsgType = "data"
	MsgTypeError                   MsgType = "err"
	MsgTypeGameCreate              MsgType = "game_created"
	MsgTypeResumeGame              MsgType = "resume_game"
	MsgTypePlayerMove              MsgType = "player_moved"
	MsgTypeMoveApproved            MsgType = "move_approved"
	MsgTypePlayerConnectionUpdated MsgType = "player_connection_updated"
	MsgTypeGameEnd                 MsgType = "game_ended"
	MsgTypeChatCreated             MsgType = "chat_created"
	MsgTypeChatMsgSend             MsgType = "msg_send"
	MsgTypeChatMsgApproved         MsgType = "msg_approved"

	MsgDataInternalErrorr string = "internal error "
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

func (m *Msg) Encode() []byte {
	b, _ := json.Marshal(m)
	return b
}
