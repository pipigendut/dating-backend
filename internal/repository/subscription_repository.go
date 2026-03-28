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
	CreateEntityBoost(ctx context.Context, boost *entities.EntityBoost) error
	GetActiveBoost(ctx context.Context, entityID uuid.UUID) (*entities.EntityBoost, error)

	// Store / Monetization
	GetAllPlans(ctx context.Context) ([]entities.SubscriptionPlan, error)
	GetPlanByID(ctx context.Context, id uuid.UUID) (*entities.SubscriptionPlan, error)
	GetConsumablePackages(ctx context.Context) ([]entities.ConsumablePackage, error)
	GetConsumablePackageByID(ctx context.Context, id uuid.UUID) (*entities.ConsumablePackage, error)
	CreateUserSubscription(ctx context.Context, sub *entities.UserSubscription) error
	AddUserConsumablePackage(ctx context.Context, userID, packageID uuid.UUID) error
}
