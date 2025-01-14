package types

type Color uint8

const (
	ColorWhite Color = iota + 1
	ColorBlack
)

func (c Color) String() string {
	if c == ColorWhite {
		return "white"
	}
	return "black"
}

type Player struct {
	ID    ObjectId
	Color Color
}
