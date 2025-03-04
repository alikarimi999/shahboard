package event

import (
	"encoding/json"

	"github.com/alikarimi999/shahboard/types"
)

var (
	TopicUsersMatchedCreated = NewTopic(DomainMatch, ActionCreated)
)

type EventUsersMatchCreated struct {
	ID        types.ObjectId `json:"id"`
	User1     types.User     `json:"user1"`
	User2     types.User     `json:"user2"`
	Timestamp int64          `json:"timestamp"`
}

func (e EventUsersMatchCreated) GetTopic() Topic {
	return TopicUsersMatchedCreated
}

func (e EventUsersMatchCreated) GetAction() Action {
	return ActionCreated
}

func (e EventUsersMatchCreated) TimeStamp() int64 {
	return e.Timestamp
}

func (e EventUsersMatchCreated) Encode() []byte {
	b, _ := json.Marshal(e)
	return b
}
