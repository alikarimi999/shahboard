package event

import (
	"encoding/json"

	"github.com/alikarimi999/shahboard/types"
	"github.com/notnil/chess"
)

const (
	ActionGamePlayerMoved              Action = "playerMoved"
	ActionGameMoveApprove              Action = "moveApproved"
	ActionGamePlayerClaimDraw          Action = "claimDraw"
	ActionGamePlayerResponsedDrawOffer Action = "playerResponsedDrawOffer"
	ActionGamePlayerClaimDrawApproved  Action = "claimDrawApproved"
	ActionGamePlayerResigned           Action = "playerResigned"
	ActionGamePlayerLeft               Action = "playerLeft"
	ActionGamePlayerJoined             Action = "playerJoined"
	ActionGamePlayerSelectSquare       Action = "selectSquare"
)

var (
	TopicGame                         = NewTopic(DomainGame, ActionAny)
	TopicGameCreated                  = NewTopic(DomainGame, ActionCreated)
	TopicGamePlayerMoved              = NewTopic(DomainGame, ActionGamePlayerMoved)
	TopicGameMoveApproved             = NewTopic(DomainGame, ActionGameMoveApprove)
	TopicGamePlayerJoined             = NewTopic(DomainGame, ActionGamePlayerJoined)
	TopicGameEnded                    = NewTopic(DomainGame, ActionEnded)
	TopicGamePlayerClaimDraw          = NewTopic(DomainGame, ActionGamePlayerClaimDraw)
	TopicGamePlayerResponsedDrawOffer = NewTopic(DomainGame, ActionGamePlayerResponsedDrawOffer)
	TopicGamePlayerClaimDrawApproved  = NewTopic(DomainGame, ActionGamePlayerClaimDrawApproved)
	TopicGamePlayerResigned           = NewTopic(DomainGame, ActionGamePlayerResigned)
	TopicGamePlayerLeft               = NewTopic(DomainGame, ActionGamePlayerLeft)
	TopicGamePlayerSelectSquare       = NewTopic(DomainGame, ActionGamePlayerSelectSquare)
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

type EventGamePlayerJoined struct {
	ID        types.ObjectId `json:"id"`
	GameID    types.ObjectId `json:"game_id"`
	PlayerID  types.ObjectId `json:"player_id"`
	Timestamp int64          `json:"timestamp"`
}

func (e EventGamePlayerJoined) GetResource() string {
	return e.GameID.String()
}

func (e EventGamePlayerJoined) GetTopic() Topic {
	return TopicGamePlayerJoined.SetResource(e.GetResource())
}

func (e EventGamePlayerJoined) GetAction() Action {
	return ActionGamePlayerJoined
}

func (e EventGamePlayerJoined) TimeStamp() int64 {
	return e.Timestamp
}

func (e EventGamePlayerJoined) Encode() []byte {
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

type EventGamePlayerClaimDraw struct {
	ID        types.ObjectId `json:"id"`
	GameID    types.ObjectId `json:"game_id"`
	PlayerID  types.ObjectId `json:"player_id"`
	Method    chess.Method   `json:"method"`
	Timestamp int64          `json:"timestamp"`
}

func (e EventGamePlayerClaimDraw) GetResource() string {
	return e.GameID.String()
}

func (e EventGamePlayerClaimDraw) GetTopic() Topic {
	return TopicGamePlayerClaimDraw.SetResource(e.GetResource())
}
func (e EventGamePlayerClaimDraw) GetAction() Action {
	return ActionGamePlayerClaimDraw
}

func (e EventGamePlayerClaimDraw) TimeStamp() int64 {
	return e.Timestamp
}

func (e EventGamePlayerClaimDraw) Encode() []byte {
	b, _ := json.Marshal(e)
	return b
}

type EventGamePlayerResponsedDrawOffer struct {
	ID        types.ObjectId `json:"id"`
	ClaimID   types.ObjectId `json:"claim_id"`
	GameID    types.ObjectId `json:"game_id"`
	PlayerID  types.ObjectId `json:"player_id"`
	Accept    bool           `json:"accept"`
	Timestamp int64          `json:"timestamp"`
}

func (e EventGamePlayerResponsedDrawOffer) GetResource() string {
	return e.GameID.String()
}

func (e EventGamePlayerResponsedDrawOffer) GetTopic() Topic {
	return TopicGamePlayerResponsedDrawOffer.SetResource(e.GetResource())
}

func (e EventGamePlayerResponsedDrawOffer) GetAction() Action {
	return ActionGamePlayerResponsedDrawOffer
}

func (e EventGamePlayerResponsedDrawOffer) TimeStamp() int64 {
	return e.Timestamp
}

func (e EventGamePlayerResponsedDrawOffer) Encode() []byte {
	b, _ := json.Marshal(e)
	return b
}

type EventGamePlayerClaimDrawApproved struct {
	ID        types.ObjectId `json:"id"`
	ClaimID   types.ObjectId `json:"claim_id"`
	GameID    types.ObjectId `json:"game_id"`
	PlayerID  types.ObjectId `json:"player_id"`
	Method    chess.Method   `json:"method"`
	Timestamp int64          `json:"timestamp"`
}

func (e EventGamePlayerClaimDrawApproved) GetResource() string {
	return e.GameID.String()
}

func (e EventGamePlayerClaimDrawApproved) GetTopic() Topic {
	return TopicGamePlayerClaimDrawApproved.SetResource(e.GetResource())
}

func (e EventGamePlayerClaimDrawApproved) GetAction() Action {
	return ActionGamePlayerClaimDraw
}

func (e EventGamePlayerClaimDrawApproved) TimeStamp() int64 {
	return e.Timestamp
}

func (e EventGamePlayerClaimDrawApproved) Encode() []byte {
	b, _ := json.Marshal(e)
	return b
}

type EventGamePlayerResigned struct {
	ID        types.ObjectId `json:"id"`
	GameID    types.ObjectId `json:"game_id"`
	PlayerID  types.ObjectId `json:"player_id"`
	Timestamp int64          `json:"timestamp"`
}

func (e EventGamePlayerResigned) GetResource() string {
	return e.GameID.String()
}

func (e EventGamePlayerResigned) GetTopic() Topic {
	return TopicGamePlayerResigned.SetResource(e.GetResource())
}

func (e EventGamePlayerResigned) GetAction() Action {
	return ActionGamePlayerResigned
}

func (e EventGamePlayerResigned) TimeStamp() int64 {
	return e.Timestamp
}

func (e EventGamePlayerResigned) Encode() []byte {
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
