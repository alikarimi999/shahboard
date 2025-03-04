package event

import (
	"encoding/json"

	"github.com/alikarimi999/shahboard/types"
)

const (
	ActionMsgSent     = "msg_sent"
	ActionMsgApproved = "msg_approved"
)

var (
	TopicGameChat            = NewTopic(DomainGameChat, ActionAny)
	TopicGameChatCreated     = NewTopic(DomainGameChat, ActionCreated)
	TopicGameChatMsgSent     = NewTopic(DomainGameChat, ActionMsgSent)
	TopicGameChatMsgApproved = NewTopic(DomainGameChat, ActionMsgApproved)
	TopicGameChatEnded       = NewTopic(DomainGameChat, ActionEnded)
)

type EventGameChatCreated struct {
	ID        types.ObjectId `json:"id"`
	GameID    types.ObjectId `json:"game_id"`
	MatchID   types.ObjectId `json:"match_id"`
	Player1   types.Player   `json:"player1"`
	Player2   types.Player   `json:"player2"`
	Timestamp int64          `json:"timestamp"`
}

func (e EventGameChatCreated) GetResource() string {
	return e.GameID.String()
}

func (e EventGameChatCreated) GetTopic() Topic {
	return TopicGameChatCreated.SetResource(e.GetResource())
}

func (e EventGameChatCreated) GetAction() Action {
	return ActionCreated
}

func (e EventGameChatCreated) TimeStamp() int64 {
	return e.Timestamp
}

func (e EventGameChatCreated) Encode() []byte {
	b, _ := json.Marshal(e)
	return b
}

type EventGameChatMsgeSent struct {
	ID        types.ObjectId `json:"id"`
	SenderID  types.ObjectId `json:"sender_id"`
	GameID    types.ObjectId `json:"game_id"`
	Content   string         `json:"content"`
	Timestamp int64          `json:"timestamp"`
}

func (e EventGameChatMsgeSent) GetResource() string {
	return e.GameID.String()
}

func (e EventGameChatMsgeSent) GetTopic() Topic {
	return TopicGameChatMsgSent.SetResource(e.GetResource())
}

func (e EventGameChatMsgeSent) GetAction() Action {
	return ActionMsgSent
}

func (e EventGameChatMsgeSent) TimeStamp() int64 {
	return e.Timestamp
}

func (e EventGameChatMsgeSent) Encode() []byte {
	b, _ := json.Marshal(e)
	return b
}

type EventGameChatMsgApproved struct {
	ID        types.ObjectId `json:"id"`
	GameID    types.ObjectId `json:"game_id"`
	SenderId  types.ObjectId `json:"sender_id"`
	Content   string         `json:"content"`
	Timestamp int64          `json:"timestamp"`
}

func (e EventGameChatMsgApproved) GetResource() string {
	return e.GameID.String()
}

func (e EventGameChatMsgApproved) GetTopic() Topic {
	return TopicGameChatMsgApproved.SetResource(e.GetResource())
}

func (e EventGameChatMsgApproved) GetAction() Action {
	return ActionMsgApproved
}

func (e EventGameChatMsgApproved) TimeStamp() int64 {
	return e.Timestamp
}

func (e EventGameChatMsgApproved) Encode() []byte {
	b, _ := json.Marshal(e)
	return b
}

type EventGameChatEnded struct {
	ID        types.ObjectId `json:"id"`
	GameID    types.ObjectId `json:"game_id"`
	Player1   types.Player   `json:"player1"`
	Player2   types.Player   `json:"player2"`
	Timestamp int64          `json:"timestamp"`
}

func (e EventGameChatEnded) GetResource() string {
	return e.GameID.String()
}

func (e EventGameChatEnded) GetTopic() Topic {
	return TopicGameChatEnded.SetResource(e.GetResource())
}

func (e EventGameChatEnded) GetAction() Action {
	return ActionEnded
}

func (e EventGameChatEnded) TimeStamp() int64 {
	return e.Timestamp
}

func (e EventGameChatEnded) Encode() []byte {
	b, _ := json.Marshal(e)
	return b
}
