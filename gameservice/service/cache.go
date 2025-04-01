package game

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/alikarimi999/shahboard/gameservice/entity"
	"github.com/alikarimi999/shahboard/pkg/log"
	"github.com/alikarimi999/shahboard/types"
	"github.com/redis/go-redis/v9"
)

const (
	keyPlayerGamePrefix  = "player_game:"
	keyGamePrefix        = "game:"
	keyLiveGamesDataHash = "live_games_data"
)

type redisGameCache struct {
	serviceID    string
	rc           *redis.Client
	deactivedTTL time.Duration
	l            log.Logger
}

func newRedisGameCache(sercviceID string, rc *redis.Client, deactivedTTL time.Duration, l log.Logger) *redisGameCache {
	return &redisGameCache{
		serviceID:    sercviceID,
		rc:           rc,
		deactivedTTL: deactivedTTL,
		l:            l,
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
func (c *redisGameCache) addGame(ctx context.Context, g *entity.Game) (bool, error) {

	script := `
	local serviceID = KEYS[1]
	local gameID = ARGV[1]
	local keyGamePrefix = KEYS[2]
	local keyPlayerGamePrefix = KEYS[3]
	local keyLiveGamesDataHash = KEYS[4]
	local gameKey = keyGamePrefix .. gameID
	local playerGameKey1 = keyPlayerGamePrefix .. ARGV[2]
	local playerGameKey2 = keyPlayerGamePrefix .. ARGV[3]
	local gameListKey = serviceID .. ":games"
	local expirationTime = tonumber(ARGV[4])

	-- MSetNX to store game data
	local success = redis.call('MSETNX', gameKey, ARGV[5])
	if success == 0 then
		return 0
	end

	if expirationTime > 0 then
	    redis.call('EXPIRE', gameKey, expirationTime)
	end

	-- Add game ID to the service's game list
	redis.call('LPUSH', gameListKey, gameID)

	-- Add player game mapping (Player1 -> GameID and Player2 -> GameID)
	redis.call('SET', playerGameKey1, gameID)
	redis.call('SET', playerGameKey2, gameID)

	-- Add the game to the live games hash
    redis.call('HSET', keyLiveGamesDataHash, gameID, ARGV[6])

	return 1
	`

	cGame := &inCacheGame{
		Status: g.Status(),
		Game:   g.Encode(),
	}

	dGame := LiveGameData{
		GameID:    g.ID(),
		Player1:   g.Player1(),
		Player2:   g.Player2(),
		StartedAt: time.Now(),
	}

	// Execute the Lua script with expiration time as an argument
	return c.rc.Eval(ctx, script, []string{
		c.serviceID,
		keyGamePrefix,
		keyPlayerGamePrefix,
		keyLiveGamesDataHash,
	}, []interface{}{
		g.ID().String(),
		g.Player1().ID.String(),
		g.Player2().ID.String(),
		0,
		cGame.encode(),
		dGame.encode(),
	}).Bool()
}

func (c *redisGameCache) playerHasGame(ctx context.Context, p types.ObjectId) (bool, error) {
	res, err := c.rc.Get(ctx, fmt.Sprintf("%s%s", keyPlayerGamePrefix, p)).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return res != "", nil
}

// need more optimizations here
func (c *redisGameCache) updateGameMove(ctx context.Context, g *entity.Game) error {
	cGame := &inCacheGame{
		Status: g.Status(),
		Game:   g.Encode(),
	}
	bGame := cGame.encode()

	return c.rc.Set(ctx, fmt.Sprintf("%s%s", keyGamePrefix, g.ID()), bGame, 0).Err()
}

func (c *redisGameCache) updateAndDeactivateGame(ctx context.Context, g *entity.Game) error {
	tx := c.rc.TxPipeline()

	// delete the players game so they can join a new game
	tx.Del(ctx, fmt.Sprintf("%s%s", keyPlayerGamePrefix, g.Player1().ID),
		fmt.Sprintf("%s%s", keyPlayerGamePrefix, g.Player2().ID))

	// remove the game from the service's game list
	tx.LRem(ctx, fmt.Sprintf("%s:games", c.serviceID), 0, g.ID().String())

	// remove the game from the live games hash
	tx.HDel(ctx, keyLiveGamesDataHash, g.ID().String())

	// set the game data with a deactivation TTL
	cGame := &inCacheGame{
		Status: g.Status(),
		Game:   g.Encode(),
	}
	bGame := cGame.encode()
	tx.Set(ctx, fmt.Sprintf("%s%s", keyGamePrefix, g.ID()), bGame, c.deactivedTTL)

	_, err := tx.Exec(ctx)
	return err
}

func (c *redisGameCache) getLiveGamesData(ctx context.Context) ([]*LiveGameData, error) {
	m, err := c.rc.HGetAll(ctx, keyLiveGamesDataHash).Result()
	if err != nil {
		return nil, err
	}

	list := []*LiveGameData{}
	for _, v := range m {
		d := &LiveGameData{}
		if err := json.Unmarshal([]byte(v), d); err != nil {
			c.l.Error(fmt.Sprintf("failed to unmarshal game data: %s", err.Error()))
			continue
		}
		list = append(list, d)
	}

	return list, nil
}

func (c *redisGameCache) getGamesIDs(ctx context.Context) ([]types.ObjectId, error) {
	maxCount := 1000
	var cursor uint64
	var gamesIDs []types.ObjectId
	for {
		keys, nextCursor, err := c.rc.Scan(ctx, cursor, fmt.Sprintf("%s*", keyGamePrefix), int64(maxCount)).Result()
		if err != nil {
			return nil, err
		}

		for _, k := range keys {
			parts := strings.Split(k, ":")
			if len(parts) != 2 {
				continue
			}
			id, err := types.ParseObjectId(parts[1])
			if err != nil {
				continue
			}
			gamesIDs = append(gamesIDs, id)
		}

		if nextCursor == 0 {
			break
		}

		cursor = nextCursor
	}

	return gamesIDs, nil
}

func (c *redisGameCache) getGameIdByUserID(ctx context.Context, userID types.ObjectId) (types.ObjectId, error) {
	gameID, err := c.rc.Get(ctx, fmt.Sprintf("%s%s", keyPlayerGamePrefix, userID)).Result()
	if err != nil {
		if err == redis.Nil {
			return types.ObjectZero, nil
		}
		return types.ObjectZero, err
	}

	return types.ParseObjectId(gameID)
}

func (c *redisGameCache) getGames(ctx context.Context) ([]*entity.Game, error) {
	maxCount := 100
	var cursor uint64
	var games []*entity.Game

	for {
		keys, nextCursor, err := c.rc.Scan(ctx, cursor, fmt.Sprintf("%s*", keyGamePrefix), int64(maxCount)).Result()
		if err != nil {
			return nil, err
		}

		gameData, err := c.rc.MGet(ctx, keys...).Result()
		if err != nil {
			return nil, err
		}

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
				c.l.Debug(fmt.Sprintf("failed to decode game: %v", err))
				continue
			}

			games = append(games, game)

		}

		if nextCursor == 0 {
			break
		}

		cursor = nextCursor
	}

	return games, nil
}

func (c *redisGameCache) getGamesByID(ctx context.Context, ids []types.ObjectId) ([]*entity.Game, error) {
	keys := make([]string, len(ids))

	for i, id := range ids {
		keys[i] = fmt.Sprintf("%s%s", keyGamePrefix, id.String())
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
			c.l.Debug(fmt.Sprintf("failed to decode game: %v", err))
			continue
		}

		gs = append(gs, game)

	}

	return gs, nil

}

func (c *redisGameCache) getGamesByServiceID(ctx context.Context, id string) ([]*entity.Game, error) {
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
			c.l.Debug(fmt.Sprintf("failed to decode game: %v", err))
			continue
		}

		gs = append(gs, game)

	}

	return gs, nil
}

func (c *redisGameCache) getGameByID(ctx context.Context, id types.ObjectId) (*entity.Game, error) {
	gameData, err := c.rc.Get(ctx, fmt.Sprintf("%s%s", keyGamePrefix, id.String())).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	if gameData == "" {
		return nil, nil
	}

	g := &inCacheGame{}
	if err := g.decode([]byte(gameData)); err != nil {
		return nil, err
	}

	game := &entity.Game{}
	if err := game.Decode(g.Game); err != nil {
		c.l.Debug(fmt.Sprintf("failed to decode game: %v", err))
		return nil, err
	}
	return game, nil
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
