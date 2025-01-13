package game

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alikarimi999/shahboard/gameservice/entity"
	"github.com/alikarimi999/shahboard/types"
	"github.com/redis/go-redis/v9"
)

const (
	keyPlayerGamePrefix = "game:player:"
	keyGamePrefix       = "game:"
)

type redisGameCache struct {
	serviceID    string
	rc           *redis.Client
	deactivedTTL time.Duration
}

func newRedisGameCache(sercviceID string, rc *redis.Client, deactivedTTL time.Duration) *redisGameCache {
	return &redisGameCache{
		serviceID:    sercviceID,
		rc:           rc,
		deactivedTTL: deactivedTTL,
	}
}

// This function uses a Lua script for atomic operations, ensuring that:
// 1. The game data is only set if it doesn't already exist (MSETNX).
// 2. The game data can have an expiration time if provided.
// 3. The game ID is added to the list of active games for the service.
// 4. Player-to-game mappings are created for both players (Player1 and Player2).
//
// Arguments:
//   - ctx: The context for the Redis operation.
//   - g: The Game object containing details like ID, players, and game data.
//   - expiration: The expiration time for the game data in seconds (0 for no expiration).
//
// Returns:
//   - bool: true if the game was successfully added, false if the game already exists.
//   - error: An error if the Redis operation fails.
func (c *redisGameCache) AddGame(ctx context.Context, g *entity.Game) (bool, error) {

	script := `
	local serviceID = KEYS[1]
	local gameID = ARGV[1]
	local keyGamePrefix = KEYS[2]
	local keyPlayerGamePrefix = KEYS[3]
	local gameKey = keyGamePrefix .. gameID
	local playerGameKey1 = keyPlayerGamePrefix .. ARGV[2]
	local playerGameKey2 = keyPlayerGamePrefix .. ARGV[3]
	local gameListKey = serviceID .. ":games"
	local expirationTime = tonumber(ARGV[4])

	-- MSetNX to store game data
	local success = redis.call('MSETNX', gameKey, ARGV[5])
	if success == 1 and expirationTime > 0 then
	    redis.call('EXPIRE', gameKey, expirationTime)
	end

	-- Add game ID to the service's game list
	redis.call('LPUSH', gameListKey, gameID)

	-- Add player game mapping (Player1 -> GameID and Player2 -> GameID)
	redis.call('SET', playerGameKey1, gameID)
	redis.call('SET', playerGameKey2, gameID)

	return success
	`

	cGame := &inCacheGame{
		Status: g.Status(),
		Game:   g.Encode(),
	}
	bGame := cGame.encode()
	// Execute the Lua script with expiration time as an argument
	return c.rc.Eval(ctx, script, []string{
		c.serviceID,
		keyGamePrefix,
		keyPlayerGamePrefix,
	}, []interface{}{
		g.ID().String(),
		g.Player1().ID.String(),
		g.Player2().ID.String(),
		0,
		bGame,
	}).Bool()
}

func (c *redisGameCache) PlayerHasGame(ctx context.Context, p types.ObjectId) (bool, error) {
	res, err := c.rc.Get(ctx, fmt.Sprintf("%s%d", keyPlayerGamePrefix, p)).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return res != "", nil
}

// need more optimizations here
func (c *redisGameCache) UpdateGameMove(ctx context.Context, g *entity.Game) error {
	cGame := &inCacheGame{
		Status: g.Status(),
		Game:   g.Encode(),
	}
	bGame := cGame.encode()

	return c.rc.Set(ctx, fmt.Sprintf("%s%d", keyGamePrefix, g.ID()), bGame, 0).Err()
}

func (c *redisGameCache) UpdateAndDeactivateGame(ctx context.Context, g *entity.Game) error {
	// delete the players game so they can join a new game
	c.rc.Del(ctx, fmt.Sprintf("%s%d", keyPlayerGamePrefix, g.Player1()))
	c.rc.Del(ctx, fmt.Sprintf("%s%d", keyPlayerGamePrefix, g.Player2()))

	cGame := &inCacheGame{
		Status: g.Status(),
		Game:   g.Encode(),
	}
	bGame := cGame.encode()

	// instead of deleting the game, we just set it to inactive and set a TTL
	return c.rc.Set(ctx, fmt.Sprintf("%s%d", keyGamePrefix, g.ID()), bGame, c.deactivedTTL).Err()
}

func (c *redisGameCache) GetGamesByServiceID(ctx context.Context, id string) ([]*entity.Game, error) {
	if id == "" {
		id = c.serviceID
	}

	gamesId, err := c.rc.LRange(ctx, fmt.Sprintf("%s:games", id), 0, -1).Result()
	if err != nil {
		return nil, err
	}

	if len(gamesId) == 0 {
		return nil, nil
	}

	keys := make([]string, len(gamesId))
	for i, gId := range gamesId {
		keys[i] = fmt.Sprintf("%s%s", keyGamePrefix, gId)
	}

	gameData, err := c.rc.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	gs := []*entity.Game{}
	for _, i := range gameData {
		if i == nil {
			continue
		}

		g := &inCacheGame{}
		if err := g.decode([]byte(i.(string))); err != nil {
			continue
		}

		game := &entity.Game{}
		if err := game.Decode(g.Game); err != nil {
			continue
		}

		gs = append(gs, game)

	}

	return gs, nil
}

type inCacheGame struct {
	Status entity.GameStatus
	Game   []byte
}

func (g *inCacheGame) encode() []byte {
	d, _ := json.Marshal(g)
	return d
}

func (g *inCacheGame) decode(data []byte) error {
	return json.Unmarshal(data, g)
}
