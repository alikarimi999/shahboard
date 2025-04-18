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
	if chance(g.b.cfg.ResignChance) {
		go func() {
			randSleep0(60, 60)
			if err := g.resign(); err != nil {
				fmt.Printf("resign error: %v\n", err)
				return
			}
			fmt.Printf("bot %s resigned from game '%s'\n", g.b.Email(), g.id)
		}()
	} else {
		go func() {
			randSleep0(60, 60)

			t := time.NewTicker(15 * time.Minute)
			<-t.C
			defer t.Stop()
			if err := g.resign(); err != nil {
				fmt.Printf("resign error: %v\n", err)
				return
			}
		}()
	}
}

func randSleep0(min, max int) {
	time.Sleep(time.Duration((rand.Intn(max))+min) * time.Second)
}

func chance(c int) bool {
	return rand.Intn(100) < c
}
