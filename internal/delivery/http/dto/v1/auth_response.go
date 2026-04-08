package dtov1

import (
)

type CheckEmailResponse struct {
	Exists bool `json:"exists" example:"true"`
}

type AuthResponse struct {
	Token        string                `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string                `json:"refresh_token" example:"abcdef123456..."`
	User         UserResponse          `json:"user"`
}

type TokenResponse struct {
	Token        string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"abcdef123456..."`
}
