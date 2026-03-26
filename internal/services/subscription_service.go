package services

import (
	"context"
	"fmt"
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
	IsBoosted(ctx context.Context, userID uuid.UUID) (bool, *time.Time, error)
	GetPlans(ctx context.Context) ([]entities.SubscriptionPlan, error)
	GetConsumableItems(ctx context.Context) ([]entities.ConsumablePackage, error)
	PurchaseConsumable(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) error
	PurchasePlan(ctx context.Context, userID uuid.UUID, planID uuid.UUID, priceID uuid.UUID) error
	GetStatus(ctx context.Context, userID uuid.UUID) (*MonetizationStatus, error)
	ActivateBoost(ctx context.Context, userID uuid.UUID) (*entities.UserBoost, error)
}

type MonetizationStatus struct {
	IsPremium   bool            `json:"is_premium"`
	PlanName    string          `json:"plan_name"`
	Features    map[string]bool `json:"features"`
	Consumables map[string]int  `json:"consumables"`
}

type subscriptionService struct {
	repo      repository.SubscriptionRepository
	userRepo  repository.UserRepository
	redisRepo repository.RedisRepository
	configSvc ConfigService
}

func NewSubscriptionService(
	repo repository.SubscriptionRepository,
	userRepo repository.UserRepository,
	redisRepo repository.RedisRepository,
	configSvc ConfigService,
) SubscriptionService {
	return &subscriptionService{
		repo:      repo,
		userRepo:  userRepo,
		redisRepo: redisRepo,
		configSvc: configSvc,
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
	counts := map[string]int{
		"boost": 0,
		"crush": 0,
	}

	consumables, err := s.repo.GetConsumables(ctx, userID)
	if err != nil {
		return counts, err
	}

	for _, cons := range consumables {
		counts[cons.ItemType] += cons.Amount
	}

	return counts, nil
}

func (s *subscriptionService) UseConsumable(ctx context.Context, userID uuid.UUID, consumableType string) (bool, error) {
	consumables, err := s.GetConsumables(ctx, userID)
	if err != nil {
		return false, err
	}

	if val, ok := consumables[consumableType]; ok && val > 0 {
		err := s.repo.UpdateConsumable(ctx, userID, consumableType, -1)
		if err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}

func (s *subscriptionService) IsBoosted(ctx context.Context, userID uuid.UUID) (bool, *time.Time, error) {
	// 1. Check Redis first
	// if s.redisRepo != nil {
	// 	expiresAt, err := s.redisRepo.GetBoostExpiration(ctx, userID)
	// 	if err == nil && expiresAt != nil {
	// 		return true, expiresAt, nil
	// 	}
	// }

	// 2. Fallback to DB to get the actual boost object
	boost, err := s.repo.GetActiveBoost(ctx, userID)
	if err != nil {
		return false, nil, err
	}

	if boost == nil {
		return false, nil, nil
	}

	// 3. Backfill Redis if active
	if s.redisRepo != nil {
		_ = s.redisRepo.SetUserBoost(ctx, userID, boost.ExpiredAt)
	}

	return true, &boost.ExpiredAt, nil
}

func (s *subscriptionService) GetPlans(ctx context.Context) ([]entities.SubscriptionPlan, error) {
	return s.repo.GetAllPlans(ctx)
}

func (s *subscriptionService) GetConsumableItems(ctx context.Context) ([]entities.ConsumablePackage, error) {
	return s.repo.GetConsumablePackages(ctx)
}

func (s *subscriptionService) PurchaseConsumable(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) error {
	return s.repo.AddUserConsumablePackage(ctx, userID, itemID)
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

func (s *subscriptionService) ActivateBoost(ctx context.Context, userID uuid.UUID) (*entities.UserBoost, error) {
	// 1. Check if user already has an active boost
	active, err := s.repo.GetActiveBoost(ctx, userID)
	if err != nil {
		return nil, err
	}
	if active != nil {
		return nil, fmt.Errorf("you already have an active boost until %s", active.ExpiredAt.Format(time.Kitchen))
	}

	// 2. Use one boost consumable
	success, err := s.UseConsumable(ctx, userID, "boost")
	if err != nil {
		return nil, err
	}
	if !success {
		return nil, fmt.Errorf("insufficient boost balance")
	}

	// 3. Create boost record
	// Default duration is 60 minutes if not specified in config
	durationMinutes := 60
	if s.configSvc != nil {
		durationMinutes = s.configSvc.GetInt("boost_duration_minutes", 60)
	}

	now := time.Now()
	expiredAt := now.Add(time.Duration(durationMinutes) * time.Minute)

	boost := &entities.UserBoost{
		UserID:    userID,
		StartedAt: now,
		ExpiredAt: expiredAt,
	}

	if err := s.repo.CreateUserBoost(ctx, boost); err != nil {
		return nil, err
	}

	// 4. Cache in Redis
	if s.redisRepo != nil {
		_ = s.redisRepo.SetUserBoost(ctx, userID, boost.ExpiredAt)
	}

	return boost, nil
}
