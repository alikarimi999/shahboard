package service

import (
	"context"
	"encoding/json"

	"github.com/alikarimi999/shahboard/chatservice/entity"
	"github.com/alikarimi999/shahboard/types"
	"github.com/redis/go-redis/v9"
)

const (
	keyGameChatPrefix = "game_chat:"
)

type redisChatCache struct {
	serviceID string
	rc        *redis.Client
}

func newRedisChatCache(serviceID string, rc *redis.Client) *redisChatCache {
	return &redisChatCache{
		serviceID: serviceID,
		rc:        rc,
	}
}

// This function uses a Lua script for atomic operations, ensuring that:
// 1. The game chat data is only set if it doesn't already exist (MSETNX).
// 2. The game chat data can have an expiration time if provided.
// 3. The game chat ID is added to the list of game chats for the service.
func (c *redisChatCache) addGameChat(ctx context.Context, gameChat *entity.Chat) (bool, error) {

	script := `
  	local serviceID = KEYS[1]
	local keyGameChatPrefix = KEYS[2]
	local gameID = ARGV[1]
	local gameChatKey = keyGameChatPrefix .. gameID
	local gameChatListKey = serviceID .. ":game_chats"
	local expirationTime = tonumber(ARGV[2])

	-- MSetNX to store game chat data
	local success = redis.call('MSETNX', gameChatKey, ARGV[3])
	if success == 0 then
		return 0
	end

	if expirationTime > 0 then
		redis.call('EXPIRE', gameChatKey, expirationTime)
	end

	-- Add game chat ID to the list of game chats
	redis.call('LPUSH', gameChatListKey, gameID)
	return 1
    `

	ic := &inCacheChatGame{
		Status: gameChat.GetStatus(),
		Chat:   gameChat.Encode(),
	}

	bChat := ic.encode()

	return c.rc.Eval(ctx, script, []string{
		c.serviceID,
		keyGameChatPrefix,
	}, []interface{}{
		gameChat.GetId().String(),
		0,
		bChat,
	}).Bool()
}

func (c *redisChatCache) deleteGameChat(ctx context.Context, gameID types.ObjectId) error {
	tx := c.rc.TxPipeline()

	tx.Del(ctx, keyGameChatPrefix+gameID.String())
	tx.LRem(ctx, c.serviceID+":game_chats", 0, gameID.String())

	_, err := tx.Exec(ctx)

	return err
}

type inCacheChatGame struct {
	Status entity.ChatStatus `json:"status"`
	Chat   []byte            `json:"chat"`
}

func (i *inCacheChatGame) encode() []byte {
	b, _ := json.Marshal(i)
	return b
}

func decodeInCacheChatGame(b []byte) (*inCacheChatGame, error) {
	var decoded inCacheChatGame
	if err := json.Unmarshal(b, &decoded); err != nil {
		return nil, err
	}
	return &decoded, nil
}
