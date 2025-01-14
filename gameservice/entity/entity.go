package entity

import (
	"fmt"
	"strings"
	"time"

	"math/rand"

	"github.com/alikarimi999/shahboard/types"
	"github.com/notnil/chess"
)

var defaultNotation = chess.AlgebraicNotation{}

type GameStatus uint8

const (
	GameStatusActive GameStatus = iota + 1
	GameStatusDeactive
)

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

type Color uint8

const (
	ColorWhite Color = 1
	ColorBlack Color = 2
)

func (c Color) String() string {
	if c == ColorWhite {
		return "white"
	}
	return "black"
}

type Player struct {
	ID    types.ObjectId
	Color Color
}

type GameSettings struct {
	Time time.Duration
}

type Game struct {
	id     types.ObjectId
	status GameStatus

	player1 Player
	player2 Player

	setting GameSettings

	game *chess.Game

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewGame(player1 types.ObjectId, player2 types.ObjectId, s GameSettings) *Game {

	p1, p2 := setPlayersId(player1, player2)
	c1, c2 := setColors()
	t := time.Now()
	g := &Game{
		id:        types.NewObjectId(),
		status:    GameStatusActive,
		player1:   Player{ID: p1, Color: c1},
		player2:   Player{ID: p2, Color: c2},
		setting:   s,
		game:      chess.NewGame(chess.UseNotation(defaultNotation)),
		CreatedAt: t,
		UpdatedAt: t,
	}

	g.game.AddTagPair("ID", g.id.String())
	g.game.AddTagPair("White", g.white().ID.String())
	g.game.AddTagPair("Black", g.black().ID.String())
	g.game.AddTagPair("created_at", t.Format(time.RFC3339))
	g.game.AddTagPair("updated_at", t.Format(time.RFC3339))

	return g
}

func (g *Game) ID() types.ObjectId {
	return g.id
}

func (g *Game) Status() GameStatus {
	return g.status
}

func (g *Game) Player1() Player {
	return g.player1
}

func (g *Game) Player2() Player {
	return g.player2
}

func (g *Game) Move(m string) error {
	if err := g.game.MoveStr(m); err != nil {
		return err
	}
	g.UpdatedAt = time.Now()
	return nil
}

func (g *Game) Turn() Player {
	if g.game.Position().Turn() == chess.White {
		return g.white()
	}
	return g.black()
}

func (g *Game) PGN() string {
	g.game.AddTagPair("updated_at", g.UpdatedAt.Format(time.RFC3339))
	return g.game.String()
}

func (g *Game) FEN() string {
	return g.game.FEN()
}

func (g *Game) white() Player {
	if g.player1.Color == ColorWhite {
		return g.player1
	}
	return g.player2
}

func (g *Game) black() Player {
	if g.player1.Color == ColorBlack {
		return g.player1
	}
	return g.player2
}

func (g *Game) Outcome() GameOutcome {
	return GameOutcome(g.game.Outcome().String())
}

func (g *Game) ValidMoves() []string {
	moves := []string{}
	for _, m := range g.game.ValidMoves() {
		moves = append(moves, defaultNotation.Encode(g.game.Position(), m))
	}
	return moves
}

func (g *Game) Deactivate() {
	g.status = GameStatusDeactive
}

func (g *Game) Encode() []byte {
	s := fmt.Sprintf("%d:%d:%d:%d:%d:%d\n", g.id, g.status,
		g.player1.ID, g.player1.Color, g.player2.ID, g.player2.Color)
	txt, _ := g.game.MarshalText()
	return []byte(s + string(txt))
}

func (g *Game) Decode(data []byte) error {

	parts := strings.SplitN(string(data), "\n", 2)
	if len(parts) < 2 {
		return fmt.Errorf("invalid encoded data")
	}

	var id, status, player1, color1, player2, color2 uint64
	_, err := fmt.Sscanf(parts[0], "%d:%d:%d:%d:%d:%d\n", &id, &status,
		&player1, &color1, &player2, &color2)
	if err != nil {
		return fmt.Errorf("failed to parse game header: %v", err)
	}

	g.id = types.ObjectId(id)
	g.status = GameStatus(status)
	g.player1 = Player{ID: types.ObjectId(player1), Color: Color(color1)}
	g.player2 = Player{ID: types.ObjectId(player2), Color: Color(color2)}

	g.game = chess.NewGame(chess.UseNotation(defaultNotation))

	if err := g.game.UnmarshalText([]byte(parts[1])); err != nil {
		return fmt.Errorf("failed to decode game text: %v", err)
	}

	return nil
}

func setPlayersId(pa, pb types.ObjectId) (p1, p2 types.ObjectId) {
	if pa < pb {
		return pa, pb
	}
	return pb, pa
}

func setColors() (Color, Color) {

	color := Color(rand.Intn(2) + 1)

	if color == ColorWhite {
		return ColorWhite, ColorBlack
	}
	return ColorBlack, ColorWhite
}
