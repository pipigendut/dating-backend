package services

import (
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/repository"
)

type AdvertisementService interface {
	GetActiveAds(placement string) ([]entities.Advertisement, error)
}

type advertisementService struct {
	repo repository.AdvertisementRepository
}

func NewAdvertisementService(repo repository.AdvertisementRepository) AdvertisementService {
	return &advertisementService{repo: repo}
}

func (s *advertisementService) GetActiveAds(placement string) ([]entities.Advertisement, error) {
	return s.repo.GetActiveAds(placement)
}
