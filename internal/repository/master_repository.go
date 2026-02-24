package repository

import "github.com/pipigendut/dating-backend/internal/entities"

type MasterRepository interface {
	GetGenders() ([]entities.MasterGender, error)
	GetRelationshipTypes() ([]entities.MasterRelationshipType, error)
	GetInterests() ([]entities.MasterInterest, error)
	GetLanguages() ([]entities.MasterLanguage, error)
}
