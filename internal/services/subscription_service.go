package services

import (
	"context"
	"time"
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/repository"
)

type SubscriptionService interface {
	GetActiveSubscription(ctx context.Context, userID uuid.UUID) (*entities.UserSubscription, error)
	HasFeature(ctx context.Context, userID uuid.UUID, featureKey string) (bool, interface{}, error)
	GetConsumables(ctx context.Context, userID uuid.UUID) (map[string]int, error)
	UseConsumable(ctx context.Context, userID uuid.UUID, consumableType string) (bool, error)
	IsBoosted(ctx context.Context, userID uuid.UUID) (bool, error)
	GetPlans(ctx context.Context) ([]entities.SubscriptionPlan, error)
	GetConsumableItems(ctx context.Context) ([]entities.ConsumableItem, error)
	PurchaseConsumable(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) error
	PurchasePlan(ctx context.Context, userID uuid.UUID, planID uuid.UUID, priceID uuid.UUID) error
	GetStatus(ctx context.Context, userID uuid.UUID) (*MonetizationStatus, error)
}

type MonetizationStatus struct {
	IsPremium   bool            `json:"is_premium"`
	PlanName    string          `json:"plan_name"`
	Features    map[string]bool `json:"features"`
	Consumables map[string]int  `json:"consumables"`
}

type subscriptionService struct {
	repo     repository.SubscriptionRepository
	userRepo repository.UserRepository
}

func NewSubscriptionService(repo repository.SubscriptionRepository, userRepo repository.UserRepository) SubscriptionService {
	return &subscriptionService{
		repo:     repo,
		userRepo: userRepo,
	}
}

func (s *subscriptionService) GetActiveSubscription(ctx context.Context, userID uuid.UUID) (*entities.UserSubscription, error) {
	return s.repo.GetActiveSubscription(ctx, userID)
}

func (s *subscriptionService) HasFeature(ctx context.Context, userID uuid.UUID, featureKey string) (bool, interface{}, error) {
	sub, err := s.GetActiveSubscription(ctx, userID)
	if err != nil {
		return false, nil, err
	}
	if sub == nil || sub.Plan == nil {
		return false, nil, nil
	}

	for _, f := range sub.Plan.Features {
		if f.FeatureKey == featureKey {
			return f.IsActive, f.IsActive, nil
		}
	}

	return false, nil, nil
}

func (s *subscriptionService) GetConsumables(ctx context.Context, userID uuid.UUID) (map[string]int, error) {
	consumables, err := s.repo.GetConsumables(ctx, userID)
	if err != nil {
		return nil, err
	}

	res := make(map[string]int)
	for _, c := range consumables {
		res[c.Type] = c.Remaining
	}
	return res, nil
}

func (s *subscriptionService) UseConsumable(ctx context.Context, userID uuid.UUID, consumableType string) (bool, error) {
	consumables, err := s.repo.GetConsumables(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, c := range consumables {
		if c.Type == consumableType && c.Remaining > 0 {
			err := s.repo.UpdateConsumable(ctx, userID, consumableType, -1)
			if err != nil {
				return false, err
			}
			return true, nil
		}
	}

	return false, nil
}

func (s *subscriptionService) IsBoosted(ctx context.Context, userID uuid.UUID) (bool, error) {
	boost, err := s.repo.GetActiveBoost(ctx, userID)
	if err != nil {
		return false, err
	}
	return boost != nil, nil
}

func (s *subscriptionService) GetPlans(ctx context.Context) ([]entities.SubscriptionPlan, error) {
	return s.repo.GetAllPlans(ctx)
}

func (s *subscriptionService) GetConsumableItems(ctx context.Context) ([]entities.ConsumableItem, error) {
	return s.repo.GetConsumableItems(ctx)
}

func (s *subscriptionService) PurchaseConsumable(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) error {
	item, err := s.repo.GetConsumableItemByID(ctx, itemID)
	if err != nil {
		return err
	}

	con := &entities.UserConsumable{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      item.ItemType,
		Remaining: item.Amount,
	}

	return s.repo.UpsertUserConsumable(ctx, con)
}

func (s *subscriptionService) PurchasePlan(ctx context.Context, userID uuid.UUID, planID uuid.UUID, priceID uuid.UUID) error {
	plan, err := s.repo.GetPlanByID(ctx, planID)
	if err != nil {
		return err
	}

	// Deactivate existing subscriptions
	// Potential future improvement: handle prorated upgrades

	// Create new subscription (simplified for demo: 30 days for any plan purchase)
	sub := &entities.UserSubscription{
		ID:        uuid.New(),
		UserID:    userID,
		PlanID:    plan.ID,
		StartedAt: time.Now(),
		ExpiredAt: time.Now().AddDate(0, 1, 0), // Default 1 month
		IsActive:  true,
	}

	if err := s.repo.CreateUserSubscription(ctx, sub); err != nil {
		return err
	}

	// Update user's premium status
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	user.IsPremium = true
	return s.userRepo.Update(user)
}

func (s *subscriptionService) GetStatus(ctx context.Context, userID uuid.UUID) (*MonetizationStatus, error) {
	sub, err := s.repo.GetActiveSubscription(ctx, userID)
	if err != nil {
		return nil, err
	}

	consumables, err := s.GetConsumables(ctx, userID)
	if err != nil {
		return nil, err
	}

	status := &MonetizationStatus{
		IsPremium:   false,
		PlanName:    "Free",
		Features:    make(map[string]bool),
		Consumables: consumables,
	}

	if sub != nil && sub.Plan != nil {
		status.IsPremium = true
		status.PlanName = sub.Plan.Name
		for _, f := range sub.Plan.Features {
			status.Features[f.FeatureKey] = f.IsActive
		}
	}

	return status, nil
}
