package bot

import (
	"encoding/json"
	"fmt"
	"net/http"

	gs "github.com/alikarimi999/shahboard/gameservice/service"
)

func (b *Bot) getLiveList() (*gs.GetLiveGamesDataResponse, error) {
	var url string
	if b.cfg.Local {
		url = b.cfg.GameService
	} else {
		url = fmt.Sprintf("%s/game", b.cfg.Server)
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/live/data", url), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", b.jwtToken))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get live game failed")
	}

	var data gs.GetLiveGamesDataResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &data, nil
}
