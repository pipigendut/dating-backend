package user

import (
	"time"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
)

type UserResponse struct {
	ID        uuid.UUID             `json:"id"`
	Email     *string               `json:"email,omitempty"`
	Status    entities.UserStatus   `json:"status"`
	Profile   *ProfileResponse      `json:"profile,omitempty"`
	Photos    []PhotoResponse       `json:"photos,omitempty"`
	CreatedAt time.Time             `json:"created_at"`
}

type ProfileResponse struct {
	FullName    string    `json:"full_name"`
	DateOfBirth time.Time `json:"date_of_birth"`
	Gender      string    `json:"gender"`
	Bio         string    `json:"bio"`
}

type PhotoResponse struct {
	URL    string `json:"url"`
	IsMain bool   `json:"is_main"`
}

func ToUserResponse(u *entities.User) UserResponse {
	resp := UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
	}

	if u.Profile != nil {
		resp.Profile = &ProfileResponse{
			FullName:    u.Profile.FullName,
			DateOfBirth: u.Profile.DateOfBirth,
			Gender:      u.Profile.Gender,
			Bio:         u.Profile.Bio,
		}
	}

	for _, p := range u.Photos {
		resp.Photos = append(resp.Photos, PhotoResponse{
			URL:    p.URL,
			IsMain: p.IsMain,
		})
	}

	return resp
}
