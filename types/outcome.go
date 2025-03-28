package types

// A Outcome is the result of a game.
type GameOutcome string

const (
	// NoOutcome indicates that a game is in progress or ended without a result.
	NoOutcome GameOutcome = "*"
	// WhiteWon indicates that white won the game.
	WhiteWon GameOutcome = "1-0"
	// BlackWon indicates that black won the game.
	BlackWon GameOutcome = "0-1"
	// Draw indicates that game was a draw.
	Draw GameOutcome = "1/2-1/2"
)

func (o GameOutcome) String() string {
	return string(o)
}

func ParseOutcome(s string) GameOutcome {
	switch s {
	case "1-0":
		return WhiteWon
	case "0-1":
		return BlackWon
	case "1/2-1/2":
		return Draw
	default:
		return NoOutcome
	}
}
