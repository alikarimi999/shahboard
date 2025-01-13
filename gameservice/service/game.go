package game

import "github.com/alikarimi999/shahboard/gameservice/entity"

type gameManager struct {
	*entity.Game
	*subscriptionManager
}

func newGameManager(gs *Service, g *entity.Game) *gameManager {
	return &gameManager{
		Game:                g,
		subscriptionManager: newSubscriptionManager(gs),
	}
}

func (gm *gameManager) stop() {
	gm.subscriptionManager.stop()
}
