package entity

import (
	"fmt"
	"strconv"
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

func (o GameOutcome) String() string {
	return string(o)
}

type GameSettings struct {
	Time time.Duration
}

type Game struct {
	id     types.ObjectId
	status GameStatus

	player1 types.Player
	player2 types.Player

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
		player1:   types.Player{ID: p1, Color: c1},
		player2:   types.Player{ID: p2, Color: c2},
		setting:   s,
		game:      chess.NewGame(chess.UseNotation(defaultNotation)),
		CreatedAt: t,
		UpdatedAt: t,
	}

	g.game.AddTagPair("id", g.id.String())
	g.game.AddTagPair("w", g.white().ID.String())
	g.game.AddTagPair("b", g.black().ID.String())
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

func (g *Game) Player1() types.Player {
	return g.player1
}

func (g *Game) Player2() types.Player {
	return g.player2
}

func (g *Game) Move(m string) error {
	if err := g.game.MoveStr(m); err != nil {
		return err
	}
	g.UpdatedAt = time.Now()
	return nil
}

func (g *Game) Turn() types.Player {
	if g.game.Position().Turn() == chess.White {
		return g.white()
	}
	return g.black()
}

func (g *Game) PGN() string {
	return g.game.String()
}

func (g *Game) FEN() string {
	return g.game.FEN()
}

func (g *Game) white() types.Player {
	if g.player1.Color == types.ColorWhite {
		return g.player1
	}
	return g.player2
}

func (g *Game) black() types.Player {
	if g.player1.Color == types.ColorBlack {
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

func (g *Game) Resign(player types.ObjectId) bool {
	return g.resign(player, EndDescriptionPlayerResigned)
}

func (g *Game) PlayerLeft(player types.ObjectId) bool {
	return g.resign(player, EndDescriptionPlayerLeft)
}

func (g *Game) EndGame() bool {
	return g.deactivate(EndDescriptionEmpty)
}

func (g *Game) resign(player types.ObjectId, desc endDescription) bool {
	if g.player1.ID == player {
		g.game.Resign(colorToChessColor(g.player1.Color))
		return g.deactivate(desc)
	} else if g.player2.ID == player {
		g.game.Resign(colorToChessColor(g.player2.Color))
		return g.deactivate(desc)
	}
	return false
}

func (g *Game) deactivate(desc endDescription) bool {
	g.status = GameStatusDeactive
	g.game.AddTagPair(endDescriptionTag, string(desc))
	g.UpdatedAt = time.Now()
	return true
}

func (g *Game) Encode() []byte {
	s := fmt.Sprintf("%s:%d:%s:%d:%s:%d\n", g.id.String(), g.status,
		g.player1.ID.String(), g.player1.Color, g.player2.ID.String(), g.player2.Color)
	txt, _ := g.game.MarshalText()
	return []byte(s + string(txt))
}

func (g *Game) Decode(data []byte) error {

	parts := strings.SplitN(string(data), "\n", 2)
	if len(parts) < 2 {
		return fmt.Errorf("invalid encoded data")
	}

	headerFields := strings.Split(parts[0], ":")
	if len(headerFields) != 6 {
		return fmt.Errorf("invalid header: expected 6 fields, got %d", len(headerFields))
	}

	id, err := types.ParseObjectId(headerFields[0])
	if err != nil {
		return fmt.Errorf("failed to parse id: %v", err)
	}

	status, err := strconv.ParseUint(headerFields[1], 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse status: %v", err)
	}

	player1, err := types.ParseObjectId(headerFields[2])
	if err != nil {
		return fmt.Errorf("failed to parse player1: %v", err)
	}

	color1, err := strconv.ParseUint(headerFields[3], 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse color1: %v", err)
	}

	player2, err := types.ParseObjectId(headerFields[4])
	if err != nil {
		return fmt.Errorf("failed to parse player2: %v", err)
	}

	color2, err := strconv.ParseUint(headerFields[5], 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse color2: %v", err)
	}

	g.id = types.ObjectId(id)
	g.status = GameStatus(status)
	g.player1 = types.Player{ID: types.ObjectId(player1), Color: types.Color(color1)}
	g.player2 = types.Player{ID: types.ObjectId(player2), Color: types.Color(color2)}

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

func setColors() (types.Color, types.Color) {

	color := types.Color(rand.Intn(2) + 1)

	if color == types.ColorWhite {
		return types.ColorWhite, types.ColorBlack
	}
	return types.ColorBlack, types.ColorWhite
}

func colorToChessColor(c types.Color) chess.Color {
	if c == types.ColorWhite {
		return chess.White
	}
	return chess.Black
}
