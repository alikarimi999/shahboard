package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

func LoadConfigs(file string, cfg interface{}) error {
	configFile, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer configFile.Close()

	if err := json.NewDecoder(configFile).Decode(cfg); err != nil {
		return fmt.Errorf("failed to decode JSON config: %w", err)
	}

	return nil
}
