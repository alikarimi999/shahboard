package event

import (
	"encoding/json"

	"github.com/alikarimi999/shahboard/types"
)

var (
	TopicGame                        = NewTopic(DomainGame, ActionAny)
	TopicGameCreated                 = NewTopic(DomainGame, ActionCreated)
	TopicGamePlayerMoved             = NewTopic(DomainGame, ActionGamePlayerMoved)
	TopicGameMoveApproved            = NewTopic(DomainGame, ActionGameMoveApprove)
	TopicGamePlayerConnectionUpdated = NewTopic(DomainGame, ActionGamePlayerConnectionUpdated)
	TopicGameEnded                   = NewTopic(DomainGame, ActionEnded)
	TopicGamePlayerLeft              = NewTopic(DomainGame, ActionGamePlayerLeft)
	TopicGamePlayerSelectSquare      = NewTopic(DomainGame, ActionGamePlayerSelectSquare)
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
	return TopicGameCreated.SetResource(e.GetResource())
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
	PlayerID  types.ObjectId `json:"player_id"`
	Move      string         `json:"move"`
	Index     int            `json:"index"`
	Timestamp int64          `json:"timestamp"`
}

func (e EventGamePlayerMoved) GetResource() string {
	return e.GameID.String()
}

func (e EventGamePlayerMoved) GetTopic() Topic {
	return TopicGamePlayerMoved.SetResource(e.GetResource())
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
	PlayerID  types.ObjectId `json:"player_id"`
	Move      string         `json:"move"`
	Index     int            `json:"index"`
	Timestamp int64          `json:"timestamp"`
}

func (e EventGameMoveApproved) GetResource() string {
	return e.GameID.String()
}

func (e EventGameMoveApproved) GetTopic() Topic {
	return TopicGameMoveApproved.SetResource(e.GetResource())
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

type EventGamePlayerConnectionUpdated struct {
	ID        types.ObjectId `json:"id"`
	GameID    types.ObjectId `json:"game_id"`
	PlayerID  types.ObjectId `json:"player_id"`
	Connected bool           `json:"connected"`
	Timestamp int64          `json:"timestamp"`
}

func (e EventGamePlayerConnectionUpdated) GetResource() string {
	return e.GameID.String()
}

func (e EventGamePlayerConnectionUpdated) GetTopic() Topic {
	return TopicGamePlayerConnectionUpdated.SetResource(e.GetResource())
}

func (e EventGamePlayerConnectionUpdated) GetAction() Action {
	return ActionGamePlayerConnectionUpdated
}

func (e EventGamePlayerConnectionUpdated) TimeStamp() int64 {
	return e.Timestamp
}

func (e EventGamePlayerConnectionUpdated) Encode() []byte {
	b, _ := json.Marshal(e)
	return b
}

type EventGameEnded struct {
	ID        types.ObjectId    `json:"id"`
	GameID    types.ObjectId    `json:"game_id"`
	Player1   types.Player      `json:"player1"`
	Player2   types.Player      `json:"player2"`
	Outcome   types.GameOutcome `json:"outcome"`
	Desc      string            `json:"desc"`
	Timestamp int64             `json:"timestamp"`
}

func (e EventGameEnded) GetResource() string {
	return e.GameID.String()
}

func (e EventGameEnded) GetTopic() Topic {
	return TopicGameEnded.SetResource(e.GetResource())
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
	PlayerID  types.ObjectId `json:"player_id"`
	Timestamp int64          `json:"timestamp"`
}

func (e EventGamePlayerLeft) GetResource() string {
	return e.GameID.String()
}

func (e EventGamePlayerLeft) GetTopic() Topic {
	return TopicGamePlayerLeft.SetResource(e.GetResource())
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

type EventGamePlayerSelectSquare struct {
	ID        types.ObjectId `json:"id"`
	GameID    types.ObjectId `json:"game_id"`
	PlayerID  types.ObjectId `json:"player_id"`
	Piece     string         `json:"piece"`
	Square    string         `json:"square"`
	Timestamp int64          `json:"timestamp"`
}

func (e EventGamePlayerSelectSquare) GetResource() string {
	return e.GameID.String()
}

func (e EventGamePlayerSelectSquare) GetTopic() Topic {
	return TopicGamePlayerSelectSquare.SetResource(e.GetResource())
}

func (e EventGamePlayerSelectSquare) GetAction() Action {
	return ActionGamePlayerSelectSquare
}

func (e EventGamePlayerSelectSquare) TimeStamp() int64 {
	return e.Timestamp
}

func (e EventGamePlayerSelectSquare) Encode() []byte {
	b, _ := json.Marshal(e)
	return b
}
