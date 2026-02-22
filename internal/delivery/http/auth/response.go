package auth

type CheckEmailResponse struct {
	Exists bool `json:"exists" example:"true"`
}

type AuthResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}
