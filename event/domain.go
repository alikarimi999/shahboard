package event

type Domain string

const (
	DomainUser       = "user"
	DomainGame       = "game"
	DomainMatch      = "match"
	DomainGameChat   = "game_chat"
	DomainDirectChat = "direct_chat"
)

func (d Domain) String() string {
	return string(d)
}
