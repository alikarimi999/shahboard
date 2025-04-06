package user

type UpdateUserRequest struct {
	Name      string `json:"name"`
	AvatarUrl string `json:"avatar_url"`
	Bio       string `json:"bio"`
	Country   string `json:"country"`
}
