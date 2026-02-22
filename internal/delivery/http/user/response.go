package user

import (
	"time"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
)

type UserResponse struct {
	ID        uuid.UUID             `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email     *string               `json:"email,omitempty" example:"user@example.com"`
	Status    entities.UserStatus   `json:"status" example:"active"`
	Profile   *ProfileResponse      `json:"profile,omitempty"`
	Photos    []PhotoResponse       `json:"photos,omitempty"`
	CreatedAt time.Time             `json:"created_at" example:"2023-01-01T00:00:00Z"`
}

type ProfileResponse struct {
	FullName    string    `json:"full_name" example:"John Doe"`
	DateOfBirth time.Time `json:"date_of_birth" example:"1995-01-01T00:00:00Z"`
	Gender      string    `json:"gender" example:"male"`
	Bio         string    `json:"bio" example:"Avid hiker and coffee lover."`
}

type PhotoResponse struct {
	ID     uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	URL    string    `json:"url" example:"https://example.com/photo.jpg"`
	IsMain bool      `json:"is_main" example:"true"`
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
			ID:     p.ID,
			URL:    p.URL,
			IsMain: p.IsMain,
		})
	}

	return resp
}
