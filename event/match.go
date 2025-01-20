package event

import (
	"encoding/json"

	"github.com/alikarimi999/shahboard/types"
)

var (
	TopicUsersMatched = NewTopic(DomainMatch, ActionPlayersMatched, "")
)

type EventUsersMatched struct {
	ID        types.ObjectId `json:"id"`
	User1     types.User     `json:"user1"`
	User2     types.User     `json:"user2"`
	Timestamp int64          `json:"timestamp"`
}

func (e EventUsersMatched) GetTopic() Topic {
	return TopicUsersMatched
}

func (e EventUsersMatched) GetAction() Action {
	return ActionPlayersMatched
}

func (e EventUsersMatched) TimeStamp() int64 {
	return e.Timestamp
}

func (e EventUsersMatched) Encode() []byte {
	b, _ := json.Marshal(e)
	return b
}
