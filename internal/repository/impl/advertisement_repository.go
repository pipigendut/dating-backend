package impl

import (
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/repository"
	"gorm.io/gorm"
)

type advertisementRepository struct {
	db *gorm.DB
}

func NewAdvertisementRepository(db *gorm.DB) repository.AdvertisementRepository {
	return &advertisementRepository{db: db}
}

func (r *advertisementRepository) GetActiveAds(placement string) ([]entities.Advertisement, error) {
	var ads []entities.Advertisement
	query := r.db.Where("is_active = ?", true)
	if placement != "" {
		query = query.Where("placement = ?", placement)
	}
	err := query.Order("sort_order ASC").Find(&ads).Error
	return ads, err
}
