package types

type Level uint8

const (
	LevelPawn Level = iota + 1
	LevelKnight
	LevelBishop
	LevelRook
	LevelQueen
	LevelKing
)
