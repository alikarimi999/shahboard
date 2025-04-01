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
	ID    ObjectId `json:"id"`
	Score int64    `json:"score"`
	Color Color    `json:"color"`
}
