package bot

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"time"

	"github.com/alikarimi999/shahboard/event"
	"github.com/alikarimi999/shahboard/types"
	"github.com/alikarimi999/shahboard/wsgateway/ws"
)

func (g *game) chatGenerator() {
	if g.color == types.ColorWhite {
		go func() {
			for {
				select {
				case <-g.stopCh:
					return
				case <-time.After(time.Second):
					if err := g.sendChat(fmt.Sprintf("random msg by white:\n '%s'\n",
						randString(defaultMsgLen))); err != nil {
						fmt.Printf("send chat error: %v\n", err)
						continue
					}
				}
			}
		}()
	} else {
		s, ok := g.subs[Topic(ws.MsgTypeChatMsgApproved)]
		if !ok {
			return
		}
		go func() {
			for {
				select {
				case <-g.stopCh:
					return
				case e := <-s.Consume():
					msgData, ok := e.Data.(*ws.Msg)
					if !ok {
						continue
					}
					msg := event.EventGameChatMsgApproved{}
					if err := json.Unmarshal(msgData.Data, &msg); err != nil {
						continue
					}
					if msg.GameID != g.id || msg.SenderId != g.opponentId {
						continue
					}

					extracted := extractMsg(msg.Content)
					if extracted == "" {
						continue
					}

					if err := g.sendChat(fmt.Sprintf("random reply to '%s' by black is:\n '%s'\n",
						extracted, randString(defaultMsgLen))); err != nil {
						fmt.Printf("send chat error: %v\n", err)
						continue
					}
				}
			}
		}()
	}
}

func randString(n int) string {
	var letterRunes = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func extractMsg(input string) string {
	re := regexp.MustCompile(`'([^']*)'`)
	match := re.FindStringSubmatch(input)

	if len(match) > 1 {
		return match[1]
	}
	return ""
}
