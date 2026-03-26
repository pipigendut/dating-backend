package user

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/delivery/http/master"
	"github.com/pipigendut/dating-backend/internal/entities"
)

type StorageURLProvider interface {
	GetPublicURL(key string) string
}

type UserSubscriptionResponse struct {
	PlanID    uuid.UUID `json:"plan_id"`
	PlanName  string    `json:"plan_name,omitempty"`
	StartedAt time.Time `json:"started_at"`
	ExpiredAt time.Time `json:"expired_at"`
	IsActive  bool      `json:"is_active"`
}

type ConsumableItemResponse struct {
	ItemType string `json:"item_type"` // boost, crush
	Amount   int    `json:"amount"`
}

type UserResponse struct {
	ID                 uuid.UUID                   `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email              *string                     `json:"email,omitempty" example:"user@example.com"`
	Status             entities.UserStatus         `json:"status" example:"active"`
	FullName           string                      `json:"full_name" example:"John Doe"`
	DateOfBirth        time.Time                   `json:"date_of_birth" example:"1995-01-01T00:00:00Z"`
	Bio                string                      `json:"bio" example:"Avid hiker and coffee lover."`
	HeightCM           int                         `json:"height_cm"`
	Gender             *master.MasterItemResponse  `json:"gender,omitempty"`
	RelationshipType   *master.MasterItemResponse  `json:"relationship_type,omitempty"`
	InterestedGenders  []master.MasterItemResponse `json:"interested_genders"`
	Interests          []master.MasterItemResponse `json:"interests"`
	Languages          []master.MasterItemResponse `json:"languages"`
	Photos             []PhotoResponse             `json:"photos"`
	LocationCity       string                      `json:"location_city,omitempty"`
	LocationCountry    string                      `json:"location_country,omitempty"`
	Latitude           *float64                    `json:"latitude,omitempty"`
	Longitude          *float64                    `json:"longitude,omitempty"`
	Age                int                         `json:"age"`
	VerifiedAt         *time.Time                  `json:"verified_at,omitempty"`
	CreatedAt          time.Time                   `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt          time.Time                   `json:"updated_at" example:"2023-01-01T00:00:00Z"`
	Subscription       *UserSubscriptionResponse   `json:"subscription,omitempty"`
	Consumables        []ConsumableItemResponse    `json:"consumables"`
}

type PhotoResponse struct {
	ID     uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	URL    string    `json:"url" example:"https://example.com/photo.jpg"`
	IsMain bool      `json:"is_main" example:"true"`
}

func ToUserResponse(u *entities.User, storage StorageURLProvider) UserResponse {
	resp := UserResponse{
		ID:                u.ID,
		Email:             u.Email,
		Status:            u.Status,
		FullName:          u.FullName,
		DateOfBirth:       u.DateOfBirth,
		Bio:               u.Bio,
		HeightCM:          u.HeightCM,
		LocationCity:      u.LocationCity,
		LocationCountry:   u.LocationCountry,
		Latitude:          u.Latitude,
		Longitude:         u.Longitude,
		Age:               u.Age,
		VerifiedAt:        u.VerifiedAt,
		CreatedAt:         u.CreatedAt,
		UpdatedAt:         u.UpdatedAt,
		InterestedGenders: make([]master.MasterItemResponse, 0),
		Interests:         make([]master.MasterItemResponse, 0),
		Languages:         make([]master.MasterItemResponse, 0),
		Photos:            make([]PhotoResponse, 0),
		Consumables:       make([]ConsumableItemResponse, 0),
	}

	if u.Gender != nil {
		resp.Gender = &master.MasterItemResponse{ID: u.Gender.ID, Name: u.Gender.Name, Icon: u.Gender.Icon}
	} else if u.GenderID != nil {
        resp.Gender = &master.MasterItemResponse{ID: *u.GenderID}
    }

	if u.RelationshipType != nil {
		resp.RelationshipType = &master.MasterItemResponse{ID: u.RelationshipType.ID, Name: u.RelationshipType.Name, Icon: u.RelationshipType.Icon}
	} else if u.RelationshipTypeID != nil {
        resp.RelationshipType = &master.MasterItemResponse{ID: *u.RelationshipTypeID}
    }

	for _, g := range u.InterestedGenders {
		resp.InterestedGenders = append(resp.InterestedGenders, master.MasterItemResponse{ID: g.ID, Name: g.Name, Icon: g.Icon})
	}
	for _, i := range u.Interests {
		resp.Interests = append(resp.Interests, master.MasterItemResponse{ID: i.ID, Name: i.Name, Icon: i.Icon})
	}
	for _, l := range u.Languages {
		resp.Languages = append(resp.Languages, master.MasterItemResponse{ID: l.ID, Name: l.Name, Icon: l.Icon})
	}

	for _, p := range u.Photos {
		url := p.URL
		if storage != nil && url != "" && !strings.HasPrefix(url, "http") {
			url = storage.GetPublicURL(url)
		}
		resp.Photos = append(resp.Photos, PhotoResponse{
			ID:     p.ID,
			URL:    url,
			IsMain: p.IsMain,
		})
	}

	// Map single active subscription
	if len(u.Subscriptions) > 0 {
		sub := u.Subscriptions[0]
		resp.Subscription = &UserSubscriptionResponse{
			PlanID:    sub.PlanID,
			StartedAt: sub.StartedAt,
			ExpiredAt: sub.ExpiredAt,
			IsActive:  sub.IsActive,
		}
		if sub.Plan != nil {
			resp.Subscription.PlanName = sub.Plan.Name
		}
	}

	// Map consumables
	if len(u.Consumables) > 0 {
		for _, cons := range u.Consumables {
			resp.Consumables = append(resp.Consumables, ConsumableItemResponse{
				ItemType: cons.ItemType,
				Amount:   cons.Amount,
			})
		}
	}

	return resp
}
