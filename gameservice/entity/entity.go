package entity

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
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

type GameSettings struct {
	Time time.Duration
}

type Game struct {
	id     types.ObjectId
	status GameStatus

	player1 types.Player
	player2 types.Player

	setting GameSettings

	lock sync.RWMutex
	game *chess.Game

	do *drawOffer

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewGame(u1 types.User, u2 types.User, s GameSettings) *Game {

	p1, p2 := setPlayersId(u1, u2)
	c1, c2 := setColors()
	p1.Color = c1
	p2.Color = c2
	t := time.Now()
	g := &Game{
		id:        types.NewObjectId(),
		status:    GameStatusActive,
		player1:   p1,
		player2:   p2,
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

func (g *Game) IsPlayer(player types.ObjectId) bool {
	return g.player1.ID == player || g.player2.ID == player
}

func (g *Game) Move(playerId types.ObjectId, move string, index int) error {
	g.lock.Lock()
	defer g.lock.Unlock()

	if g.turn().ID != playerId {
		return fmt.Errorf("it's not your turn")
	}

	if (index - 1) != len(g.game.Moves()) {
		return fmt.Errorf("invalid move index")
	}

	if err := g.game.MoveStr(move); err != nil {
		return err
	}
	g.UpdatedAt = time.Now()
	return nil
}

func (g *Game) turn() types.Player {
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

// TODO: implement some rules and limitations for sending draw offer by players
func (g *Game) OfferDraw(offerer types.ObjectId) bool {
	if !g.IsPlayer(offerer) {
		return false
	}

	g.lock.Lock()
	defer g.lock.Unlock()

	// player can't offer draw if it's not his turn
	if g.turn().ID != offerer {
		return false
	}

	// the offer should not proccess if another offer is already in progress
	if g.do != nil {
		return false
	}

	g.do = &drawOffer{
		offerer:   offerer,
		method:    chess.DrawOffer,
		timestamp: time.Now(),
	}

	return true
}

func (g *Game) AcceptDraw(acceptor types.ObjectId) bool {
	if !g.IsPlayer(acceptor) {
		return false
	}

	g.lock.Lock()
	defer g.lock.Unlock()

	if g.do == nil {
		return false
	}

	if acceptor == g.do.offerer {
		return false
	}

	if err := g.game.Draw(chess.DrawOffer); err != nil {
		return false
	}

	g.do.accepted = true

	return true
}

func (g *Game) RejectDraw(rejector types.ObjectId) bool {
	if !g.IsPlayer(rejector) {
		return false
	}

	g.lock.Lock()
	defer g.lock.Unlock()
	if g.do == nil {
		return false
	}

	if rejector == g.do.offerer {
		return false
	}

	g.do = nil
	return true
}

// only accept ThreefoldRepetition and FiftyMoveRule draw methods
func (g *Game) ClaimDraw(playerId types.ObjectId, method chess.Method) bool {
	if method != chess.ThreefoldRepetition && method != chess.FiftyMoveRule {
		return false
	}

	if !g.IsPlayer(playerId) {
		return false
	}

	g.lock.Lock()
	defer g.lock.Unlock()

	if err := g.game.Draw(method); err != nil {
		return false
	}

	return true
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

func (g *Game) Outcome() types.GameOutcome {
	return types.GameOutcome(g.game.Outcome().String())
}

func (g *Game) ValidMoves() []string {
	moves := []string{}
	for _, m := range g.game.ValidMoves() {
		moves = append(moves, defaultNotation.Encode(g.game.Position(), m))
	}
	return moves
}

// return false if player is not in game
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

func setPlayersId(u1, u2 types.User) (p1, p2 types.Player) {
	if u1.ID < u2.ID {
		return types.Player{ID: u1.ID, Score: u1.Score}, types.Player{ID: u2.ID, Score: u2.Score}
	}
	return types.Player{ID: u2.ID, Score: u2.Score}, types.Player{ID: u1.ID, Score: u1.Score}
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

type drawOffer struct {
	offerer   types.ObjectId
	method    chess.Method
	accepted  bool
	timestamp time.Time
}
