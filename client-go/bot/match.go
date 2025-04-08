package bot

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alikarimi999/shahboard/event"
	gh "github.com/alikarimi999/shahboard/gameservice/delivery/http"
	gs "github.com/alikarimi999/shahboard/gameservice/service"
	"github.com/alikarimi999/shahboard/types"
)

func (b *Bot) FindMatch() (event.EventUsersMatchCreated, error) {
	e := event.EventUsersMatchCreated{}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/match/find", b.url), nil)
	if err != nil {
		return e, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", b.jwtToken))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return e, err
	}

	if res.StatusCode != http.StatusOK {
		return e, fmt.Errorf("find match failed")
	}

	if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
		return e, err
	}

	return e, nil
}

func (b *Bot) GetUserLiveGame(userId types.ObjectId) (types.ObjectId, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/game/live/user/%s", b.url, userId), nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", b.jwtToken))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get live game failed")
	}

	g := gh.GetLiveGameIdByUserIdRequest{}
	if err := json.NewDecoder(res.Body).Decode(&g); err != nil {
		return "", err
	}

	return g.GameId, nil
}

func (b *Bot) GetLivePgnByUserId(id types.ObjectId) (gs.GetGamePGNResponse, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/game/live?user_id=%s", b.url, id), nil)
	if err != nil {
		return gs.GetGamePGNResponse{}, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", b.jwtToken))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return gs.GetGamePGNResponse{}, err
	}

	if res.StatusCode != http.StatusOK {
		return gs.GetGamePGNResponse{}, fmt.Errorf("get game pgn failed")
	}

	var g gs.GetGamePGNResponse
	if err := json.NewDecoder(res.Body).Decode(&g); err != nil {
		return gs.GetGamePGNResponse{}, err
	}

	return g, nil
}
