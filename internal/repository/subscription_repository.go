package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
)

type SubscriptionRepository interface {
	GetActiveSubscription(ctx context.Context, userID uuid.UUID) (*entities.UserSubscription, error)
	GetPlanFeatures(ctx context.Context, planID uuid.UUID) ([]entities.SubscriptionPlanFeature, error)
	GetConsumables(ctx context.Context, userID uuid.UUID) ([]entities.UserConsumable, error)
	UpdateConsumable(ctx context.Context, userID uuid.UUID, consumableType string, delta int) error
	CreateUserBoost(ctx context.Context, boost *entities.UserBoost) error
	GetActiveBoost(ctx context.Context, userID uuid.UUID) (*entities.UserBoost, error)

	// Store / Monetization
	GetAllPlans(ctx context.Context) ([]entities.SubscriptionPlan, error)
	GetPlanByID(ctx context.Context, id uuid.UUID) (*entities.SubscriptionPlan, error)
	GetConsumableItems(ctx context.Context) ([]entities.ConsumableItem, error)
	GetConsumableItemByID(ctx context.Context, id uuid.UUID) (*entities.ConsumableItem, error)
	CreateUserSubscription(ctx context.Context, sub *entities.UserSubscription) error
	UpsertUserConsumable(ctx context.Context, con *entities.UserConsumable) error
}
