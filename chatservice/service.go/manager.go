package chat

import (
	"sync"

	"github.com/alikarimi999/shahboard/chatservice/entity"
	"github.com/alikarimi999/shahboard/types"
)

type chatsManager struct {
	mu    sync.RWMutex
	chats map[types.ObjectId]*entity.Chat
}

func newChatsManager() *chatsManager {
	return &chatsManager{
		chats: make(map[types.ObjectId]*entity.Chat),
	}
}

// createChat creates a chat for a game with the provided owners.
// It returns the created chat.
// If a chat already exists for the game, it returns nil.
func (m *chatsManager) createChat(gameId types.ObjectId, player1, player2 types.Player) *entity.Chat {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.chats[gameId]; ok {
		return nil
	}

	chat := entity.NewChat(gameId, player1, player2)
	m.chats[gameId] = chat

	return chat
}

// getChat returns the chat for the game with the provided ID.
// It returns nil if no chat exists for the game.
func (m *chatsManager) getChat(gameId types.ObjectId) *entity.Chat {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.chats[gameId]
}

func (m *chatsManager) exists(gameId types.ObjectId) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, ok := m.chats[gameId]
	return ok
}

// removeChat removes the chat for the game with the provided ID.
func (m *chatsManager) removeChat(gameId types.ObjectId) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.chats, gameId)
}
