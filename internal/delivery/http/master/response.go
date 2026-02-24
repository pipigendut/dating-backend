package master

import (
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/usecases"
)

type MasterItemResponse struct {
	ID   uuid.UUID `json:"id" example:"bd0a597a-2d88-44e2-a0b4-eb416c1f2115"`
	Name string    `json:"name" example:"Travel"`
	Icon string    `json:"icon" example:"✈️"`
}

func ToMasterItemsResponseGenders(genders []usecases.GenderDTO) []MasterItemResponse {
	var results []MasterItemResponse
	for _, g := range genders {
		results = append(results, MasterItemResponse{ID: g.ID, Name: g.Name, Icon: g.Icon})
	}
	return results
}

func ToMasterItemsResponseRelationshipTypes(types []usecases.RelationshipTypeDTO) []MasterItemResponse {
	var results []MasterItemResponse
	for _, t := range types {
		results = append(results, MasterItemResponse{ID: t.ID, Name: t.Name, Icon: t.Icon})
	}
	return results
}

func ToMasterItemsResponseInterests(interests []usecases.InterestDTO) []MasterItemResponse {
	var results []MasterItemResponse
	for _, i := range interests {
		results = append(results, MasterItemResponse{ID: i.ID, Name: i.Name, Icon: i.Icon})
	}
	return results
}

func ToMasterItemsResponseLanguages(languages []usecases.LanguageDTO) []MasterItemResponse {
	var results []MasterItemResponse
	for _, l := range languages {
		results = append(results, MasterItemResponse{ID: l.ID, Name: l.Name, Icon: l.Icon})
	}
	return results
}
