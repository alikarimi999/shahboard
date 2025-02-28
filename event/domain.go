package event

type Domain string

const (
	DomainUser     = "user"
	DomainGame     = "game"
	DomainMatch    = "match"
	DomainGameChat = "game_chat"
)

func (d Domain) String() string {
	return string(d)
}
