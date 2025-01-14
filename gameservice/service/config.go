package game

import (
	"fmt"

	"github.com/alikarimi999/shahboard/gameservice/entity"
)

type Config struct {
	InstanceID          string              `json:"instance_id"`
	GamesCap            uint64              `json:"games_cap"`
	DefaultGameSettings entity.GameSettings `json:"default_game_settings"`
}

func (cfg Config) Validate() error {
	if cfg.InstanceID == "" {
		return fmt.Errorf("instance id is required")
	}

	if cfg.GamesCap == 0 {
		return fmt.Errorf("games cap is required")
	}

	if cfg.DefaultGameSettings.Time == 0 {
		return fmt.Errorf("default game setting time is required")
	}

	return nil
}
