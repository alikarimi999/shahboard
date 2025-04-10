package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	auth "github.com/alikarimi999/shahboard/authservice/service"
	profile "github.com/alikarimi999/shahboard/profileservice/service/user"
	"github.com/alikarimi999/shahboard/types"
)

// return true if user was created account before and logged in
func (b *Bot) Login() (bool, error) {
	bReq, err := json.Marshal(auth.PasswordAuthRequest{
		Email:    b.email,
		Password: b.password,
	})
	if err != nil {
		return false, err
	}

	var url string
	if b.cfg.Local {
		url = b.cfg.AuthService
	} else {
		url = fmt.Sprintf("%s/auth", b.cfg.Server)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/", url), bytes.NewBuffer(bReq))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}

	if res.StatusCode != http.StatusOK {
		return false, fmt.Errorf("login failed")
	}

	var resBody auth.PasswordAuthResponse
	if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
		return false, err
	}

	b.id = types.ObjectId(resBody.Id)
	b.jwtToken = resBody.JwtToken
	return resBody.Exists, nil
}

func (b *Bot) UpdateProfile(name, avatar string) error {
	bReq, err := json.Marshal(profile.UpdateUserRequest{
		Name:      name,
		AvatarUrl: avatar,
	})
	if err != nil {
		return err
	}

	var url string
	if b.cfg.Local {
		url = b.cfg.ProfileService
	} else {
		url = fmt.Sprintf("%s/profile", b.cfg.Server)
	}

	req, err := http.NewRequest("PATCH", fmt.Sprintf("%s/users/", url), bytes.NewBuffer(bReq))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", b.jwtToken))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("update profile failed")
	}

	return nil
}
