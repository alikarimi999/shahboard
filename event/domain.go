package event

type Domain string

const (
	DomainGame  = "game"
	DomainMatch = "match"
)

func (d Domain) String() string {
	return string(d)
}
