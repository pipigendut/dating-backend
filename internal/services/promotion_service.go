package services

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/repository"
)

type PromotionService interface {
	HandleRegistrationPromotion(ctx context.Context, userID uuid.UUID) error
}

type promotionService struct {
	subscriptionRepo repository.SubscriptionRepository
	userRepo         repository.UserRepository
	configSvc        ConfigService
}

func NewPromotionService(
	subscriptionRepo repository.SubscriptionRepository,
	userRepo repository.UserRepository,
	configSvc ConfigService,
) PromotionService {
	return &promotionService{
		subscriptionRepo: subscriptionRepo,
		userRepo:         userRepo,
		configSvc:        configSvc,
	}
}

func (s *promotionService) HandleRegistrationPromotion(ctx context.Context, userID uuid.UUID) error {
	// 1. Check if promotion is enabled
	isPromoEnabled := s.configSvc.GetString("register_promotion", "false") == "true"
	if !isPromoEnabled {
		return nil
	}

	log.Printf("[PromotionService] Applying registration promotion for user: %s", userID)

	// 2. Grant 3 Profile Boosts
	err := s.subscriptionRepo.UpdateConsumable(ctx, userID, "boost", 3)
	if err != nil {
		log.Printf("[PromotionService] Error granting boosts: %v", err)
	}

	// 3. Grant 10 Crushes
	err = s.subscriptionRepo.UpdateConsumable(ctx, userID, "crush", 10)
	if err != nil {
		log.Printf("[PromotionService] Error granting crushes: %v", err)
	}

	// 4. Grant 1 Month Ultimate Subscription
	// Ultimate Plan ID from seeds/master_data.go
	ultimatePlanID := uuid.MustParse("d0000000-0000-0000-0000-000000000003")
	
	sub := &entities.UserSubscription{
		UserID:    userID,
		PlanID:    ultimatePlanID,
		StartedAt: time.Now(),
		ExpiredAt: time.Now().AddDate(0, 1, 0), // 1 month
		IsActive:  true,
	}

	if err := s.subscriptionRepo.CreateUserSubscription(ctx, sub); err != nil {
		log.Printf("[PromotionService] Error creating ultimate subscription: %v", err)
		return err
	}

	// 5. Update User Premium Status
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		log.Printf("[PromotionService] Error fetching user to update premium status: %v", err)
		return err
	}
	
	user.IsPremium = true
	if err := s.userRepo.Update(user); err != nil {
		log.Printf("[PromotionService] Error updating user premium status: %v", err)
		return err
	}

	log.Printf("[PromotionService] Successfully applied registration promotion for user: %s", userID)

	return nil
}
