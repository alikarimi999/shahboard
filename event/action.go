package event

type Action string

const (
	ActionAny                         Action = ""
	ActionCreated                     Action = "created"
	ActionEnded                       Action = "ended"
	ActionUpdated                     Action = "updated"
	ActionDeleted                     Action = "deleted"
	ActionGamePlayerMoved             Action = "playerMoved"
	ActionGameMoveApprove             Action = "moveApproved"
	ActionGamePlayerLeft              Action = "playerLeft"
	ActionGamePlayerConnectionUpdated Action = "playerConnectionUpdated"
	ActionUsersMatched                Action = "usersMatched"
	ActionGamePlayerSelectSquare      Action = "selectSquare"
)

func (a Action) String() string {
	return string(a)
}
