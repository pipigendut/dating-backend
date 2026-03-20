package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/repository"
)

type AdminService interface {
	SubscribeUser(ctx context.Context, userID uuid.UUID, planID uuid.UUID) (*entities.User, error)
	AddConsumable(ctx context.Context, userID, packageID uuid.UUID) (*entities.User, error)
	GetFullUser(ctx context.Context, userID uuid.UUID) (*entities.User, error)
}

type adminService struct {
	repo     repository.SubscriptionRepository
	userRepo repository.UserRepository
}

func NewAdminService(repo repository.SubscriptionRepository, userRepo repository.UserRepository) AdminService {
	return &adminService{
		repo:     repo,
		userRepo: userRepo,
	}
}

func (s *adminService) SubscribeUser(ctx context.Context, userID uuid.UUID, planID uuid.UUID) (*entities.User, error) {
	sub := &entities.UserSubscription{
		UserID:    userID,
		PlanID:    planID,
		StartedAt: time.Now(),
		ExpiredAt: time.Now().AddDate(0, 1, 0), // Default 1 month for simulation
		IsActive:  true,
	}

	if err := s.repo.CreateUserSubscription(ctx, sub); err != nil {
		return nil, err
	}

	// Update user's premium status surgically to avoid wiping associations
	if err := s.userRepo.UpdatePremiumStatus(userID, true); err != nil {
		return nil, err
	}
	
	return s.GetFullUser(ctx, userID)
}

func (s *adminService) AddConsumable(ctx context.Context, userID, packageID uuid.UUID) (*entities.User, error) {
	err := s.repo.AddUserConsumablePackage(ctx, userID, packageID)
	if err != nil {
		return nil, err
	}

	return s.GetFullUser(ctx, userID)
}

func (s *adminService) GetFullUser(ctx context.Context, userID uuid.UUID) (*entities.User, error) {
	return s.userRepo.GetWithRelations(userID)
}
