package response

import (
	"strings"

	"github.com/pipigendut/dating-backend/internal/entities"
)

type StorageURLProvider interface {
	GetPublicURL(key string) string
}

func ToUserResponse(u *entities.User, storage StorageURLProvider) UserResponse {
	resp := UserResponse{
		ID:                u.ID,
		EntityID:          u.EntityID,
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
		InterestedGenders: make([]MasterItemResponse, 0),
		Interests:         make([]MasterItemResponse, 0),
		Languages:         make([]MasterItemResponse, 0),
		Photos:            make([]PhotoResponse, 0),
		Consumables:       make([]ConsumableItemResponse, 0),
	}

	if u.Gender != nil {
		resp.Gender = &MasterItemResponse{ID: u.Gender.ID, Name: u.Gender.Name, Icon: u.Gender.Icon}
	} else if u.GenderID != nil {
		resp.Gender = &MasterItemResponse{ID: *u.GenderID}
	}

	if u.RelationshipType != nil {
		resp.RelationshipType = &MasterItemResponse{ID: u.RelationshipType.ID, Name: u.RelationshipType.Name, Icon: u.RelationshipType.Icon}
	} else if u.RelationshipTypeID != nil {
		resp.RelationshipType = &MasterItemResponse{ID: *u.RelationshipTypeID}
	}

	for _, g := range u.InterestedGenders {
		resp.InterestedGenders = append(resp.InterestedGenders, MasterItemResponse{ID: g.ID, Name: g.Name, Icon: g.Icon})
	}
	for _, i := range u.Interests {
		resp.Interests = append(resp.Interests, MasterItemResponse{ID: i.ID, Name: i.Name, Icon: i.Icon})
	}
	for _, l := range u.Languages {
		resp.Languages = append(resp.Languages, MasterItemResponse{ID: l.ID, Name: l.Name, Icon: l.Icon})
	}

	for _, p := range u.Photos {
		url := p.URL
		if storage != nil && url != "" && !strings.HasPrefix(url, "http") {
			url = storage.GetPublicURL(url)
		}
		photoResp := PhotoResponse{
			ID:     p.ID,
			URL:    url,
			IsMain: p.IsMain,
		}
		resp.Photos = append(resp.Photos, photoResp)

		if p.IsMain {
			resp.MainPhoto = url
		}
	}

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

func ToUserLiteResponse(u *entities.User, storage StorageURLProvider) UserResponse {
	resp := ToUserResponse(u, storage)

	// Exclude specifically requested fields for "lite" profile
	resp.InterestedGenders = nil
	resp.Interests = nil
	resp.Languages = nil
	resp.Consumables = nil
	resp.Photos = nil

	return resp
}
