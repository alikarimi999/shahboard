package event

import (
	"encoding/json"

	"github.com/alikarimi999/shahboard/types"
)

var (
	TopicMatch Topic = NewTopic("match", "")
)

type EventPlayersMatched struct {
	Player1   types.ObjectId `json:"player1"`
	Player2   types.ObjectId `json:"player2"`
	Timestamp int64          `json:"timestamp"`
}

func (e EventPlayersMatched) GetTopic() Topic {
	return TopicMatch
}

func (e EventPlayersMatched) GetAction() Action {
	return ActionPlayersMatched
}

func (e EventPlayersMatched) TimeStamp() int64 {
	return e.Timestamp
}

func (e EventPlayersMatched) Encode() []byte {
	b, _ := json.Marshal(e)
	return b
}
