package dtov1

import (
	"github.com/google/uuid"
)

type MatchResponse struct {
	IsMatch       bool                     `json:"is_match"`
	MatchID       uuid.UUID                `json:"match_id,omitempty"`
	MatchedEntity *EntityResponse `json:"matched_entity,omitempty"`
}

type IncomingLikeResponse struct {
	Entity         EntityResponse `json:"entity"`
	IsCrush        bool                    `json:"is_crush"`
	IsBoosted      bool                    `json:"is_boosted"`
	SwipeTime      string                  `json:"swipe_time"`
	TargetEntityID string                  `json:"target_entity_id"`
}

type SentLikeResponse struct {
	Entity         EntityResponse `json:"entity"`
	IsCrush        bool                    `json:"is_crush"`
	IsBoosted      bool                    `json:"is_boosted"`
	CreatedAt      string                  `json:"created_at"`
	ExpiresAt      string                  `json:"expires_at"`
	SwiperEntityID string                  `json:"swiper_entity_id"`
}

type LikesSummaryResponse struct {
	Count     int    `json:"count"`
	LastPhoto string `json:"last_photo"`
}
