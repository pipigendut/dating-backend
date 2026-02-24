package impl

import (
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/repository"
	"gorm.io/gorm"
)

type masterRepository struct {
	db *gorm.DB
}

func NewMasterRepository(db *gorm.DB) repository.MasterRepository {
	return &masterRepository{db: db}
}

func (r *masterRepository) GetGenders() ([]entities.MasterGender, error) {
	var results []entities.MasterGender
	err := r.db.Where("is_active = ?", true).Order("name asc").Find(&results).Error
	return results, err
}

func (r *masterRepository) GetRelationshipTypes() ([]entities.MasterRelationshipType, error) {
	var results []entities.MasterRelationshipType
	err := r.db.Where("is_active = ?", true).Order("name asc").Find(&results).Error
	return results, err
}

func (r *masterRepository) GetInterests() ([]entities.MasterInterest, error) {
	var results []entities.MasterInterest
	err := r.db.Where("is_active = ?", true).Order("name asc").Find(&results).Error
	return results, err
}

func (r *masterRepository) GetLanguages() ([]entities.MasterLanguage, error) {
	var results []entities.MasterLanguage
	err := r.db.Where("is_active = ?", true).Order("name asc").Find(&results).Error
	return results, err
}
