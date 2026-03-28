package response

import (
	"time"
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
)

type MasterItemResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Icon string    `json:"icon"`
}

type UserSubscriptionResponse struct {
	PlanID    uuid.UUID `json:"plan_id"`
	PlanName  string    `json:"plan_name,omitempty"`
	StartedAt time.Time `json:"started_at"`
	ExpiredAt time.Time `json:"expired_at"`
	IsActive  bool      `json:"is_active"`
}

type ConsumableItemResponse struct {
	ItemType string `json:"item_type"` 
	Amount   int    `json:"amount"`
}

type PhotoResponse struct {
	ID     uuid.UUID `json:"id"`
	URL    string    `json:"url"`
	IsMain bool      `json:"is_main"`
}

type UserResponse struct {
	ID                 uuid.UUID                   `json:"id"`
	EntityID           uuid.UUID                   `json:"entity_id"`
	Email              *string                     `json:"email,omitempty"`
	Status             entities.UserStatus         `json:"status"`
	FullName           string                      `json:"full_name"`
	DateOfBirth        time.Time                   `json:"date_of_birth"`
	Bio                string                      `json:"bio"`
	HeightCM           int                         `json:"height_cm"`
	Gender             *MasterItemResponse         `json:"gender,omitempty"`
	RelationshipType   *MasterItemResponse         `json:"relationship_type,omitempty"`
	InterestedGenders  []MasterItemResponse        `json:"interested_genders,omitempty"`
	Interests          []MasterItemResponse        `json:"interests,omitempty"`
	Languages          []MasterItemResponse        `json:"languages,omitempty"`
	Photos             []PhotoResponse             `json:"photos,omitempty"`
	LocationCity       string                      `json:"location_city,omitempty"`
	LocationCountry    string                      `json:"location_country,omitempty"`
	Latitude           *float64                    `json:"latitude,omitempty"`
	Longitude          *float64                    `json:"longitude,omitempty"`
	Age                int                         `json:"age"`
	MainPhoto          string                      `json:"main_photo,omitempty"`
	VerifiedAt         *time.Time                  `json:"verified_at,omitempty"`
	CreatedAt          time.Time                   `json:"created_at"`
	UpdatedAt          time.Time                   `json:"updated_at"`
	Subscription       *UserSubscriptionResponse   `json:"subscription,omitempty"`
	Consumables        []ConsumableItemResponse    `json:"consumables,omitempty"`
}

type EntityResponse struct {
	ID    uuid.UUID      `json:"id"`
	Type  string         `json:"type"` 
	User  *UserResponse  `json:"user,omitempty"`
	Group *GroupResponse `json:"group,omitempty"`
}

type GroupResponse struct {
	ID        uuid.UUID      `json:"id"`
	EntityID  uuid.UUID      `json:"entity_id"`
	Name      string         `json:"name"`
	CreatedBy uuid.UUID      `json:"created_by"`
	Members   []UserResponse `json:"members,omitempty"`
}



type VerificationResult struct {
	IsMatch    bool    `json:"is_match"`
	Confidence float64 `json:"confidence"`
	Error      string  `json:"error,omitempty"`
}
