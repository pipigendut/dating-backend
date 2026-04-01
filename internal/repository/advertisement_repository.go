package repository

import "github.com/pipigendut/dating-backend/internal/entities"

type AdvertisementRepository interface {
	GetActiveAds(placement string) ([]entities.Advertisement, error)
}
