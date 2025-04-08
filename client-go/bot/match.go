package bot

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alikarimi999/shahboard/event"
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
