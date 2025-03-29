package entity

import (
	"time"

	"github.com/alikarimi999/shahboard/types"
)

type Rating struct {
	UserId       types.ObjectId
	CurrentScore int64
	BestScore    int64
	GamesPlayed  int64
	GamesWon     int64
	GamesLost    int64
	GamesDraw    int64
	LastUpdated  time.Time
}

type GameResult int8

const (
	GameResultLoss GameResult = iota - 1
	GameResultDraw
	GameResultWin
)

type GameEloChange struct {
	Id         int64
	UserId     types.ObjectId
	EloChange  int64
	GameId     types.ObjectId
	OpponentId types.ObjectId
	Result     GameResult
	UpdatedAt  time.Time
}
