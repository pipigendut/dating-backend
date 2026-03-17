package swipe

import (
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
)

type SwipeRequest struct {
	SwipedID  uuid.UUID               `json:"swiped_id" binding:"required"`
	Direction entities.SwipeDirection `json:"direction" binding:"required,oneof=LIKE DISLIKE CRUSH" swaggertype:"string"`
}

type MatchResponse struct {
	IsMatch bool      `json:"is_match"`
	MatchID uuid.UUID `json:"match_id,omitempty"`
}

// UserSwipeProfileResponse is what the client sees for swiping
// Note: It's typically same as User response, but might have less sensitive info
type UserSwipeProfileResponse struct {
	ID              uuid.UUID  `json:"id"`
	FullName        string     `json:"full_name"`
	Age             int        `json:"age"`
	Bio             string     `json:"bio"`
	HeightCM        int        `json:"height_cm"`
	LocationCity    string     `json:"location_city"`
	LocationCountry string     `json:"location_country"`
	Photos          []PhotoDTO `json:"photos"`
	// Additional fields like interests can be mapped here
}

type IncomingLikeResponse struct {
	User          UserSwipeProfileResponse `json:"user"`
	IsCrush       bool                     `json:"is_crush"`
	PriorityScore int                      `json:"priority_score"`
	SwipeTime     string                   `json:"swipe_time"`
}

type PhotoDTO struct {
	ID        uuid.UUID `json:"id"`
	URL       string    `json:"url"`
	IsMain    bool      `json:"is_main"`
	SortOrder int       `json:"sort_order"`
}
