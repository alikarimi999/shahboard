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
	TopicUser          = NewTopic(DomainUser, ActionAny)
	TopicUserCreated   = NewTopic(DomainUser, ActionCreated)
	TopicUserUpdated   = NewTopic(DomainUser, ActionUpdated)
	TopicUserDeleted   = NewTopic(DomainUser, ActionDeleted)
	TopicUserLoggedIn  = NewTopic(DomainUser, ActionLoggedIn)
	TopicUserLoggedOut = NewTopic(DomainUser, ActionLoggedOut)
)

type EventUserCreated struct {
	ID        types.ObjectId `json:"id"`
	UserID    types.ObjectId `json:"user_id"`
	IsGuest   bool           `json:"is_guest"`
	Email     string         `json:"email"`
	Name      string         `json:"name"`
	Picture   string         `json:"picture"`
	Timestamp int64          `json:"timestamp"`
}

func (e EventUserCreated) GetResource() string {
	return e.UserID.String()
}

func (e EventUserCreated) GetTopic() Topic {
	return TopicUserCreated.SetResource(e.GetResource())
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
	return TopicUserLoggedIn.SetResource(e.GetResource())
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
	return TopicUserLoggedOut.SetResource(e.GetResource())
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
