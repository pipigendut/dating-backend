package user

import (
	"time"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
)

type UserResponse struct {
	ID              uuid.UUID           `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email           *string             `json:"email,omitempty" example:"user@example.com"`
	Status          entities.UserStatus `json:"status" example:"active"`
	FullName        string              `json:"full_name" example:"John Doe"`
	DateOfBirth     time.Time           `json:"date_of_birth" example:"1995-01-01T00:00:00Z"`
	Bio             string              `json:"bio" example:"Avid hiker and coffee lover."`
	HeightCM        int                 `json:"height_cm"`
	Gender          *string             `json:"gender,omitempty"`
	LookingFor      []string            `json:"looking_for,omitempty"`
	InterestedIn    []string            `json:"interested_in,omitempty"`
	Interests       []string            `json:"interests,omitempty"`
	Languages       []string            `json:"languages,omitempty"`
	Photos          []PhotoResponse     `json:"photos,omitempty"`
	LocationCity    string              `json:"location_city,omitempty"`
	LocationCountry string              `json:"location_country,omitempty"`
	Latitude        *float64            `json:"latitude,omitempty"`
	Longitude       *float64            `json:"longitude,omitempty"`
	CreatedAt       time.Time           `json:"created_at" example:"2023-01-01T00:00:00Z"`
}

type PhotoResponse struct {
	ID     uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	URL    string    `json:"url" example:"https://example.com/photo.jpg"`
	IsMain bool      `json:"is_main" example:"true"`
}

func ToUserResponse(u *entities.User) UserResponse {
	resp := UserResponse{
		ID:              u.ID,
		Email:           u.Email,
		Status:          u.Status,
		FullName:        u.FullName,
		DateOfBirth:     u.DateOfBirth,
		Bio:             u.Bio,
		HeightCM:        u.HeightCM,
		LocationCity:    u.LocationCity,
		LocationCountry: u.LocationCountry,
		Latitude:        u.Latitude,
		Longitude:       u.Longitude,
		CreatedAt:       u.CreatedAt,
	}

	if u.GenderID != nil {
		g := u.GenderID.String()
		resp.Gender = &g
	}
	if u.RelationshipTypeID != nil {
		resp.LookingFor = append(resp.LookingFor, u.RelationshipTypeID.String())
	}

	for _, g := range u.InterestedGenders {
		resp.InterestedIn = append(resp.InterestedIn, g.ID.String())
	}
	for _, i := range u.Interests {
		resp.Interests = append(resp.Interests, i.ID.String())
	}
	for _, l := range u.Languages {
		resp.Languages = append(resp.Languages, l.ID.String())
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
