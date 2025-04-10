package bot

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/alikarimi999/shahboard/types"
	"github.com/alikarimi999/shahboard/wsgateway/ws"
)

type viewManager struct {
	email string
	mu    sync.RWMutex
	list  map[types.ObjectId]struct{}
}

func NewViewManager(email string) *viewManager {
	return &viewManager{
		email: email,
		list:  make(map[types.ObjectId]struct{}),
	}
}

func (v *viewManager) add(id types.ObjectId) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.list[id] = struct{}{}
	fmt.Printf("bot '%s' add game %s to view list\n", v.email, id)
}

func (v *viewManager) remove(id types.ObjectId) {
	v.mu.Lock()
	defer v.mu.Unlock()
	_, ok := v.list[id]
	if ok {
		delete(v.list, id)
		fmt.Printf("bot '%s' remove game %s from view list\n", v.email, id)
	}
}

func (v *viewManager) exists(id types.ObjectId) bool {
	v.mu.RLock()
	defer v.mu.RUnlock()

	_, ok := v.list[id]
	return ok
}

func (b *Bot) RandomView() {
	t := time.NewTicker(time.Second * time.Duration(rand.Intn(10)+30))

	for {
		select {
		case <-t.C:
			res, err := b.getLiveList()
			if err != nil {
				fmt.Printf("get live game failed: %v\n", err)
				continue
			}

			if len(res.List) == 0 {
				continue
			}

			for {
				// choose a random game
				game := res.List[rand.Intn(len(res.List))]
				if b.vm.exists(game.GameID) {
					continue
				}

				if err := b.SendWsMessage(ws.Msg{
					MsgBase: ws.MsgBase{
						Type:      ws.MsgTypeViewGame,
						Timestamp: time.Now().Unix(),
					},
					Data: ws.DataGameViewRequest{
						GameId: game.GameID,
					}.Encode(),
				}); err != nil {
					fmt.Printf("send ws message failed: %v\n", err)
				}
				break
			}
		case <-b.stopCh:
			t.Stop()
			return
		}
	}
}
