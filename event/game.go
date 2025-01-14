package event

import (
	"encoding/json"

	"github.com/alikarimi999/shahboard/types"
)

var (
	TopicGame Topic = NewTopic(DomainGame, "{gameID}")
)

type EventGameCreated struct {
	ID        types.ObjectId `json:"id"`
	Player1   types.Player   `json:"player1"`
	Player2   types.Player   `json:"player2"`
	Timestamp int64          `json:"timestamp"`
}

func (e EventGameCreated) GetTopic() Topic {
	t := TopicGame.WithResource(e.ID.String())
	return t
}

func (e EventGameCreated) GetAction() Action {
	return ActionCreated
}

func (e EventGameCreated) TimeStamp() int64 {
	return e.Timestamp
}

func (e EventGameCreated) Encode() []byte {
	b, _ := json.Marshal(e)
	return b
}

type EventGamePlayerMoved struct {
	ID        types.ObjectId `json:"id"`
	PlayerID  types.ObjectId `json:"playerId"`
	Move      string         `json:"move"`
	Timestamp int64          `json:"timestamp"`
}

func (e EventGamePlayerMoved) GetTopic() Topic {
	t := TopicGame.WithResource(e.ID.String())
	return t
}

func (e EventGamePlayerMoved) GetAction() Action {
	return ActionGamePlayerMoved
}

func (e EventGamePlayerMoved) TimeStamp() int64 {
	return e.Timestamp
}

func (e EventGamePlayerMoved) Encode() []byte {
	b, _ := json.Marshal(e)
	return b
}

type EventGameMoveApproved struct {
	ID        types.ObjectId `json:"id"`
	PlayerID  types.ObjectId `json:"playerId"`
	Move      string         `json:"move"`
	Timestamp int64          `json:"timestamp"`
}

func (e EventGameMoveApproved) GetTopic() Topic {
	t := TopicGame.WithResource(e.ID.String())
	return t
}

func (e EventGameMoveApproved) GetAction() Action {
	return ActionGameMoveApprove
}

func (e EventGameMoveApproved) TimeStamp() int64 {
	return e.Timestamp
}

func (e EventGameMoveApproved) Encode() []byte {
	b, _ := json.Marshal(e)
	return b
}

type EventGameEnded struct {
	ID        types.ObjectId `json:"id"`
	Player1   types.ObjectId `json:"player1"`
	Player2   types.ObjectId `json:"player2"`
	Desc      string         `json:"desc"`
	Timestamp int64          `json:"timestamp"`
}

func (e EventGameEnded) GetTopic() Topic {
	t := TopicGame.WithResource(e.ID.String())
	return t
}

func (e EventGameEnded) GetAction() Action {
	return ActionEnded
}

func (e EventGameEnded) TimeStamp() int64 {
	return e.Timestamp
}

func (e EventGameEnded) Encode() []byte {
	b, _ := json.Marshal(e)
	return b
}

type EventGamePlayerLeft struct {
	ID        types.ObjectId `json:"id"`
	PlayerID  types.ObjectId `json:"playerId"`
	Timestamp int64          `json:"timestamp"`
}

func (e EventGamePlayerLeft) GetTopic() Topic {
	t := TopicGame.WithResource(e.ID.String())
	return t
}

func (e EventGamePlayerLeft) GetAction() Action {
	return ActionGamePlayerLeft
}

func (e EventGamePlayerLeft) TimeStamp() int64 {
	return e.Timestamp
}

func (e EventGamePlayerLeft) Encode() []byte {
	b, _ := json.Marshal(e)
	return b
}
