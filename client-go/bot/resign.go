package bot

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/alikarimi999/shahboard/event"
	"github.com/alikarimi999/shahboard/wsgateway/ws"
)

func (g *game) resign() error {
	t := time.Now().Unix()
	return g.b.SendWsMessage(ws.Msg{
		MsgBase: ws.MsgBase{
			Type:      ws.MsgTypePlayerResigned,
			Timestamp: time.Now().Unix(),
		},
		Data: ws.DataGamePlayerResignRequest{
			EventGamePlayerResigned: event.EventGamePlayerResigned{
				GameID:    g.id,
				PlayerID:  g.b.ID(),
				Timestamp: t,
			},
		}.Encode(),
	})
}

func (g *game) resigner() {
	if resignChance() {
		go func() {
			randSleep(60)
			if err := g.resign(); err != nil {
				fmt.Printf("resign error: %v\n", err)
				return
			}
			fmt.Printf("bot %s resigned from game '%s'\n", g.b.Email(), g.id)
		}()
	}
}

func resignChance() bool {
	return rand.Intn(100) < 10
}
