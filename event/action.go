package event

type Action string

const (
	ActionAny             Action = ""
	ActionCreated         Action = "created"
	ActionEnded           Action = "ended"
	ActionGamePlayerMoved Action = "playerMoved"
	ActionGameMoveApprove Action = "moveApproved"
	ActionGamePlayerLeft  Action = "playerLeft"
	ActionPlayersMatched  Action = "playersMatched"
)

func (a Action) String() string {
	return string(a)
}
