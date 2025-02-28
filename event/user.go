package event

import (
	"encoding/json"

	"github.com/alikarimi999/shahboard/types"
)

const (
	ActionLoggedIn  = "loggedIn"
	ActionLoggedOut = "loggedOut"
)

var (
	TopicUser          = NewTopic(DomainUser, ActionAny, "{userID}")
	TopicUserCreated   = NewTopic(DomainUser, ActionCreated, "{user}")
	TopicUserUpdated   = NewTopic(DomainUser, ActionUpdated, "{user}")
	TopicUserDeleted   = NewTopic(DomainUser, ActionDeleted, "{user}")
	TopicUserLoggedIn  = NewTopic(DomainUser, ActionLoggedIn, "{user}")
	TopicUserLoggedOut = NewTopic(DomainUser, ActionLoggedOut, "{user}")
)

type EventUserCreated struct {
	ID        types.ObjectId `json:"id"`
	UserID    types.ObjectId `json:"user_id"`
	Email     string         `json:"email"`
	Name      string         `json:"name"`
	Picture   string         `json:"picture"`
	Timestamp int64          `json:"timestamp"`
}

func (e EventUserCreated) GetResource() string {
	return e.UserID.String()
}

func (e EventUserCreated) GetTopic() Topic {
	return TopicUser.WithResource(e.GetResource())
}

func (e EventUserCreated) GetAction() Action {
	return ActionCreated
}

func (e EventUserCreated) TimeStamp() int64 {
	return e.Timestamp
}

func (e EventUserCreated) Encode() []byte {
	b, _ := json.Marshal(e)
	return b
}

type EventUserLoggedIn struct {
	ID        types.ObjectId `json:"id"`
	UserID    types.ObjectId `json:"user_id"`
	Email     string         `json:"email"`
	Timestamp int64          `json:"timestamp"`
}

func (e EventUserLoggedIn) GetResource() string {
	return e.UserID.String()
}

func (e EventUserLoggedIn) GetTopic() Topic {
	return TopicUser.WithResource(e.GetResource())
}

func (e EventUserLoggedIn) GetAction() Action {
	return ActionLoggedIn
}
func (e EventUserLoggedIn) TimeStamp() int64 {
	return e.Timestamp
}
func (e EventUserLoggedIn) Encode() []byte {
	b, _ := json.Marshal(e)
	return b
}

type EventUserLoggedOut struct {
	ID        types.ObjectId `json:"id"`
	UserID    types.ObjectId `json:"user_id"`
	Email     string         `json:"email"`
	Timestamp int64          `json:"timestamp"`
}

func (e EventUserLoggedOut) GetResource() string {
	return e.UserID.String()
}

func (e EventUserLoggedOut) GetTopic() Topic {
	return TopicUser.WithResource(e.GetResource())
}

func (e EventUserLoggedOut) GetAction() Action {
	return ActionLoggedOut
}

func (e EventUserLoggedOut) TimeStamp() int64 {
	return e.Timestamp
}
func (e EventUserLoggedOut) Encode() []byte {
	b, _ := json.Marshal(e)
	return b
}
