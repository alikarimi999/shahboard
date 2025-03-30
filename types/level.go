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

func (l Level) String() string {
	switch l {
	case LevelPawn:
		return "pawn"
	case LevelKnight:
		return "knight"
	case LevelBishop:
		return "bishop"
	case LevelRook:
		return "rook"
	case LevelQueen:
		return "queen"
	case LevelKing:
		return "king"
	default:
		return "pawn"
	}
}
