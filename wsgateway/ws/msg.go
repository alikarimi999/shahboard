package ws

import (
	"encoding/json"

	"github.com/alikarimi999/shahboard/types"
)

type MsgType string

const (
	MsgTypeWelcome         MsgType = "welcome"
	MsgTypeFindMatch       MsgType = "find_match"
	MsgTypeViewGame        MsgType = "view_game"
	MsgTypeData            MsgType = "data"
	MsgTypeError           MsgType = "err"
	MsgTypeGameCreate      MsgType = "game_created"
	MsgTypeResumeGame      MsgType = "resume_game"
	MsgTypePlayerMove      MsgType = "player_moved"
	MsgTypeMoveApproved    MsgType = "move_approved"
	MsgTypePlayerJoined    MsgType = "player_joined"
	MsgTypePlayerLeft      MsgType = "player_left"
	MsgTypeGameEnd         MsgType = "game_ended"
	MsgTypeChatCreated     MsgType = "chat_created"
	MsgTypeChatMsgSend     MsgType = "msg_send"
	MsgTypeChatMsgApproved MsgType = "msg_approved"
	MsgTypeViewersList     MsgType = "viewers_list"

	MsgTypePlayerResigned MsgType = "player_resigned"

	MsgDataInternalErrorr string = "internal error"
	MsgDataBadRequest     string = "bad request"
	MsgDataNotFound       string = "not found"
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

func (s *session) handleMsg(sess *session, msg *Msg) {
	switch msg.Type {
	case MsgTypeFindMatch:
		var d DataFindMatchRequest
		if err := json.Unmarshal(msg.Data, &d); err != nil {
			sess.sendErr(msg.ID, "invalid data")
			return
		}

		sess.handleFindMatchRequest(msg.ID, d)
	case MsgTypeResumeGame:
		var d DataResumeGameRequest
		if err := json.Unmarshal(msg.Data, &d); err != nil {
			sess.sendErr(msg.ID, "invalid data")
			return
		}

		sess.handleResumeGameRequest(msg.ID, d)
	case MsgTypeViewGame:
		var d DataGameViewRequest
		if err := json.Unmarshal(msg.Data, &d); err != nil {
			sess.sendErr(msg.ID, "invalid data")
			return
		}

		sess.handleViewGameRequest(msg.ID, d)
	case MsgTypePlayerMove:
		var d DataGamePlayerMoveRequest
		if err := json.Unmarshal(msg.Data, &d); err != nil {
			sess.sendErr(msg.ID, "invalid data")
			return
		}

		sess.handleMoveRequest(msg.ID, d)
	case MsgTypePlayerResigned:
		var d DataGamePlayerResignRequest
		if err := json.Unmarshal(msg.Data, &d); err != nil {
			sess.sendErr(msg.ID, "invalid data")
			return
		}

		sess.handlePlayerResignRequest(msg.ID, d)
	case MsgTypeChatMsgSend:
		var d DataGameChatMsgSend
		if err := json.Unmarshal(msg.Data, &d); err != nil {
			sess.sendErr(msg.ID, "invalid data")
			return
		}

		sess.handleSendMsg(msg.ID, d)
	case MsgTypeData:

		// handle data message
	}
}
