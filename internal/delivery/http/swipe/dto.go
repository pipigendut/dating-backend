package swipe

import (
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/delivery/http/user"
)

type SwipeRequest struct {
	SwipedID  uuid.UUID               `json:"swiped_id" binding:"required"`
	Direction entities.SwipeDirection `json:"direction" binding:"required,oneof=LIKE DISLIKE CRUSH" swaggertype:"string"`
}

type MatchResponse struct {
	IsMatch     bool               `json:"is_match"`
	MatchID     uuid.UUID          `json:"match_id,omitempty"`
	MatchedUser *user.UserResponse `json:"matched_user,omitempty"`
}

type IncomingLikeResponse struct {
	User          user.UserResponse `json:"user"`
	IsCrush       bool              `json:"is_crush"`
	PriorityScore int               `json:"priority_score"`
	SwipeTime     string            `json:"swipe_time"`
}
