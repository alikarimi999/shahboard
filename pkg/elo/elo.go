package elo

import (
	"math"

	"github.com/alikarimi999/shahboard/types"
)

const (
	KNew          = 32
	KIntermediate = 24
	KExpert       = 16

	MinRating  = 1000
	BaseRating = 1000
)

func CalculateElo(s1, s2 int64, score float64) int64 {
	if s1 < BaseRating {
		s1 = BaseRating
	}
	if s2 < BaseRating {
		s2 = BaseRating
	}

	expectedScore := 1 / (1 + math.Pow(10, float64(s2-s1)/400))

	k := calculateKFactor(s1)

	newRating := float64(s1) + float64(k)*(score-expectedScore)
	if newRating < float64(MinRating) {
		return MinRating
	}
	return int64(newRating)
}

func calculateKFactor(elo int64) int64 {
	switch GetPlayerLevel(elo) {
	case types.LevelPawn, types.LevelKnight:
		return KNew
	case types.LevelBishop, types.LevelRook:
		return KIntermediate
	case types.LevelQueen, types.LevelKing:
		return KExpert
	default:
		return KExpert
	}
}

func GetPlayerLevel(elo int64) types.Level {
	switch {
	case elo <= 1200:
		return types.LevelPawn
	case elo <= 1400:
		return types.LevelKnight
	case elo <= 1600:
		return types.LevelBishop
	case elo <= 1800:
		return types.LevelRook
	case elo <= 2000:
		return types.LevelQueen
	default:
		return types.LevelKing
	}
}
