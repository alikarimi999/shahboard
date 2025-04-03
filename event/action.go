package event

type Action string

const (
	ActionAny     Action = "*"
	ActionCreated Action = "created"
	ActionEnded   Action = "ended"
	ActionUpdated Action = "updated"
	ActionDeleted Action = "deleted"
)

func (a Action) String() string {
	return string(a)
}
