package swipe

import (
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/delivery/http/response"
)

type SwipeRequest struct {
	SwiperEntityID uuid.UUID               `json:"swiper_entity_id" binding:"required"`
	SwipedEntityID uuid.UUID               `json:"swiped_entity_id" binding:"required"`
	Direction      entities.SwipeDirection `json:"direction" binding:"required,oneof=LIKE DISLIKE CRUSH" swaggertype:"string"`
}

type MatchResponse struct {
	IsMatch       bool                    `json:"is_match"`
	MatchID       uuid.UUID               `json:"match_id,omitempty"`
	MatchedEntity *response.EntityResponse `json:"matched_entity,omitempty"`
}

type IncomingLikeResponse struct {
	Entity         response.EntityResponse `json:"entity"`
	IsCrush        bool                    `json:"is_crush"`
	IsBoosted      bool                    `json:"is_boosted"`
	SwipeTime      string                  `json:"swipe_time"`
	TargetEntityID string                  `json:"target_entity_id"`
}

type SentLikeResponse struct {
	Entity         response.EntityResponse `json:"entity"`
	IsCrush        bool                    `json:"is_crush"`
	IsBoosted      bool                    `json:"is_boosted"`
	CreatedAt      string                  `json:"created_at"`
	ExpiresAt      string                  `json:"expires_at"`
	SwiperEntityID string                  `json:"swiper_entity_id"`
}

type UnlikeRequest struct {
	TargetEntityID uuid.UUID `json:"target_entity_id" binding:"required"`
}

type SwipeCandidatesFilter struct {
	Distance          *int     `json:"distance" form:"distance"`
	MinAge            *int     `json:"min_age" form:"min_age"`
	MaxAge            *int     `json:"max_age" form:"max_age"`
	Genders           []string `json:"genders" form:"genders"`
	Interests         []string `json:"interests" form:"interests"`
	RelationshipTypes []string `json:"relationship_types" form:"relationship_types"`
	Latitude          *float64 `json:"latitude" form:"latitude"`
	Longitude         *float64 `json:"longitude" form:"longitude"`
	MinHeight         *int     `json:"min_height" form:"min_height"`
	MaxHeight         *int     `json:"max_height" form:"max_height"`
	SwiperEntityID    string   `form:"swiper_entity_id" binding:"required"`
	EntityType        string   `json:"entity_type" form:"entity_type"` // "user" or "group", empty = all
}

type LikesSummaryResponse struct {
	Count     int    `json:"count"`
	LastPhoto string `json:"last_photo"`
}
