package event

import (
	"encoding/json"

	"github.com/alikarimi999/shahboard/types"
)

var (
	TopicMatch Topic = NewTopic("match", "")
)

type EventUsersMatched struct {
	ID        types.ObjectId `json:"id"`
	User1     types.User     `json:"user1"`
	User2     types.User     `json:"user2"`
	Timestamp int64          `json:"timestamp"`
}

func (e EventUsersMatched) GetTopic() Topic {
	return TopicMatch
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
