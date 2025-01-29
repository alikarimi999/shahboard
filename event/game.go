package event

import (
	"encoding/json"

	"github.com/alikarimi999/shahboard/types"
)

var (
	TopicGame             = NewTopic(DomainGame, ActionAny, "")
	TopicGameCreated      = NewTopic(DomainGame, ActionCreated, "{gameID}")
	TopicGamePlayerMoved  = NewTopic(DomainGame, ActionGamePlayerMoved, "{gameID}")
	TopicGameMoveApproved = NewTopic(DomainGame, ActionGameMoveApprove, "{gameID}")
	TopicGameEnded        = NewTopic(DomainGame, ActionEnded, "{gameID}")
	TopicGamePlayerLeft   = NewTopic(DomainGame, ActionGamePlayerLeft, "{gameID}")
)

type EventGameCreated struct {
	ID        types.ObjectId `json:"id"`
	GameID    types.ObjectId `json:"game_id"`
	MatchID   types.ObjectId `json:"match_id"`
	Player1   types.Player   `json:"player1"`
	Player2   types.Player   `json:"player2"`
	Timestamp int64          `json:"timestamp"`
}

func (e EventGameCreated) GetResource() string {
	return e.GameID.String()
}

func (e EventGameCreated) GetTopic() Topic {
	return TopicGameCreated.WithResource(e.GetResource())
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
	GameID    types.ObjectId `json:"game_id"`
	PlayerID  types.ObjectId `json:"playerId"`
	Move      string         `json:"move"`
	Timestamp int64          `json:"timestamp"`
}

func (e EventGamePlayerMoved) GetResource() string {
	return e.GameID.String()
}

func (e EventGamePlayerMoved) GetTopic() Topic {
	return TopicGamePlayerMoved.WithResource(e.GetResource())
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
	GameID    types.ObjectId `json:"game_id"`
	PlayerID  types.ObjectId `json:"playerId"`
	Move      string         `json:"move"`
	Timestamp int64          `json:"timestamp"`
}

func (e EventGameMoveApproved) GetResource() string {
	return e.GameID.String()
}

func (e EventGameMoveApproved) GetTopic() Topic {
	return TopicGameMoveApproved.WithResource(e.GetResource())
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
	GameID    types.ObjectId `json:"game_id"`
	Player1   types.Player   `json:"player1"`
	Player2   types.Player   `json:"player2"`
	Outcome   string         `json:"outcome"`
	Desc      string         `json:"desc"`
	Timestamp int64          `json:"timestamp"`
}

func (e EventGameEnded) GetResource() string {
	return e.GameID.String()
}

func (e EventGameEnded) GetTopic() Topic {
	return TopicGameEnded.WithResource(e.GetResource())
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
	GameID    types.ObjectId `json:"game_id"`
	PlayerID  types.ObjectId `json:"playerId"`
	Timestamp int64          `json:"timestamp"`
}

func (e EventGamePlayerLeft) GetResource() string {
	return e.GameID.String()
}

func (e EventGamePlayerLeft) GetTopic() Topic {
	return TopicGamePlayerLeft.WithResource(e.GetResource())
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
