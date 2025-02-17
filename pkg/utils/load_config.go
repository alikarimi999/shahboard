package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

func LoadConfigs(service string, development bool, cfg interface{}) error {
	var file string
	if development {
		file = fmt.Sprintf("deploy/%s/development/config.json", service)
	} else {
		file = fmt.Sprintf("deploy/%s/production/config.json", service)
	}
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
