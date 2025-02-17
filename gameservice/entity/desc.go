package entity

const endDescriptionTag = "end_description"

type endDescription string

const (
	EndDescriptionEmpty          endDescription = ""
	EndDescriptionPlayerResigned endDescription = "player_resigned"
	EndDescriptionPlayerLeft     endDescription = "player_left"
	EndDescriptionPlayerTimeout  endDescription = "player_timeout"
	EndDescriptionGameTimeout    endDescription = "game_timeout"
)

func (e endDescription) String() string {
	return string(e)
}
