package service

type GoogleAuthRequest struct {
	Token string `json:"token"`
}

type GoogleAuthResponse struct {
	Id       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Picture  string `json:"picture"`
	JwtToken string `json:"jwt_token"`
	Exists   bool   `json:"exists"`
}

type PasswordAuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type PasswordAuthResponse struct {
	Id       string `json:"id"`
	Email    string `json:"email"`
	JwtToken string `json:"jwt_token"`
	Exists   bool   `json:"exists"`
}

type GuestLoginResponse struct {
	Id       string `json:"id"`
	JwtToken string `json:"jwt_token"`
}
