package impl

import (
	"context"
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type subscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) repository.SubscriptionRepository {
	return &subscriptionRepository{db: db}
}

func (r *subscriptionRepository) GetActiveSubscription(ctx context.Context, userID uuid.UUID) (*entities.UserSubscription, error) {
	var sub entities.UserSubscription
	err := r.db.WithContext(ctx).
		Preload("Plan").
		Preload("Plan.Features").
		Where("user_id = ? AND is_active = true AND expired_at > ?", userID, time.Now()).
		First(&sub).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &sub, nil
}

func (r *subscriptionRepository) GetPlanFeatures(ctx context.Context, planID uuid.UUID) ([]entities.SubscriptionPlanFeature, error) {
	var features []entities.SubscriptionPlanFeature
	err := r.db.WithContext(ctx).
		Where("plan_id = ?", planID).
		Find(&features).Error
	return features, err
}

func (r *subscriptionRepository) GetConsumables(ctx context.Context, userID uuid.UUID) ([]entities.UserConsumable, error) {
	var consumables []entities.UserConsumable
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND (expired_at IS NULL OR expired_at > ?)", userID, time.Now()).
		Find(&consumables).Error
	return consumables, err
}

func (r *subscriptionRepository) UpdateConsumable(ctx context.Context, userID uuid.UUID, consumableType string, delta int) error {
	return r.db.WithContext(ctx).
		Model(&entities.UserConsumable{}).
		Where("user_id = ? AND type = ?", userID, consumableType).
		Update("remaining", gorm.Expr("remaining + ?", delta)).Error
}

func (r *subscriptionRepository) CreateUserBoost(ctx context.Context, boost *entities.UserBoost) error {
	return r.db.WithContext(ctx).Create(boost).Error
}

func (r *subscriptionRepository) GetActiveBoost(ctx context.Context, userID uuid.UUID) (*entities.UserBoost, error) {
	var boost entities.UserBoost
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_active = true AND expired_at > ?", userID, time.Now()).
		Order("expired_at DESC").
		First(&boost).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &boost, nil
}

func (r *subscriptionRepository) GetAllPlans(ctx context.Context) ([]entities.SubscriptionPlan, error) {
	var plans []entities.SubscriptionPlan
	err := r.db.WithContext(ctx).
		Preload("Features").
		Preload("Prices").
		Where("is_active = true").
		Find(&plans).Error
	return plans, err
}

func (r *subscriptionRepository) GetConsumableItems(ctx context.Context) ([]entities.ConsumableItem, error) {
	var items []entities.ConsumableItem
	err := r.db.WithContext(ctx).Find(&items).Error
	return items, err
}

func (r *subscriptionRepository) GetPlanByID(ctx context.Context, id uuid.UUID) (*entities.SubscriptionPlan, error) {
	var plan entities.SubscriptionPlan
	err := r.db.WithContext(ctx).Preload("Features").Preload("Prices").First(&plan, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *subscriptionRepository) GetConsumableItemByID(ctx context.Context, id uuid.UUID) (*entities.ConsumableItem, error) {
	var item entities.ConsumableItem
	err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *subscriptionRepository) CreateUserSubscription(ctx context.Context, sub *entities.UserSubscription) error {
	return r.db.WithContext(ctx).Create(sub).Error
}

func (r *subscriptionRepository) UpsertUserConsumable(ctx context.Context, con *entities.UserConsumable) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "type"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"remaining":  gorm.Expr("user_consumables.remaining + ?", con.Remaining),
			"updated_at": time.Now(),
		}),
	}).Create(con).Error
}
