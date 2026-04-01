package impl

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
		Where("user_id = ? AND amount > 0", userID).
		Find(&consumables).Error
	return consumables, err
}

func (r *subscriptionRepository) UpdateConsumable(ctx context.Context, userID uuid.UUID, consumableType string, delta int) error {
	now := time.Now()
	cons := entities.UserConsumable{
		UserID:   userID,
		ItemType: consumableType,
		Amount:   delta,
	}

	if delta < 0 {
		// If decreasing, we MUST check existing balance first
		return r.db.Transaction(func(tx *gorm.DB) error {
			var existing entities.UserConsumable
			if err := tx.Where("user_id = ? AND item_type = ?", userID, consumableType).First(&existing).Error; err != nil {
				return err
			}
			if existing.Amount+delta < 0 {
				return gorm.ErrRecordNotFound // Insufficient
			}
			return tx.Model(&existing).Updates(map[string]interface{}{
				"amount":       existing.Amount + delta,
				"last_used_at": &now,
			}).Error
		})
	}

	// If increasing (like for promotions), we use Upsert (OnConflict)
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "item_type"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"amount":     gorm.Expr("user_consumables.amount + ?", delta),
			"updated_at": now,
		}),
	}).Create(&cons).Error
}

func (r *subscriptionRepository) CreateEntityBoost(ctx context.Context, boost *entities.EntityBoost) error {
	return r.db.WithContext(ctx).Create(boost).Error
}

func (r *subscriptionRepository) GetActiveBoost(ctx context.Context, entityID uuid.UUID) (*entities.EntityBoost, error) {
	var boost entities.EntityBoost
	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("entity_id = ? AND started_at <= ? AND expires_at > ?", entityID, now, now).
		First(&boost).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No active boost
		}
		return nil, err
	}
	return &boost, nil
}

// Store / Monetization

func (r *subscriptionRepository) GetAllPlans(ctx context.Context) ([]entities.SubscriptionPlan, error) {
	var plans []entities.SubscriptionPlan
	err := r.db.WithContext(ctx).
		Preload("Features").
		Preload("Prices").
		Where("is_active = ?", true).
		Find(&plans).Error
	return plans, err
}

func (r *subscriptionRepository) GetPlanByID(ctx context.Context, id uuid.UUID) (*entities.SubscriptionPlan, error) {
	var plan entities.SubscriptionPlan
	err := r.db.WithContext(ctx).
		Preload("Features").
		Preload("Prices").
		Where("id = ? AND is_active = ?", id, true).
		First(&plan).Error
	return &plan, err
}

func (r *subscriptionRepository) GetConsumablePackages(ctx context.Context) ([]entities.ConsumablePackage, error) {
	var packages []entities.ConsumablePackage
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Find(&packages).Error
	return packages, err
}

func (r *subscriptionRepository) GetConsumablePackageByID(ctx context.Context, id uuid.UUID) (*entities.ConsumablePackage, error) {
	var pkg entities.ConsumablePackage
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_active = ?", id, true).
		First(&pkg).Error
	return &pkg, err
}

func (r *subscriptionRepository) CreateUserSubscription(ctx context.Context, sub *entities.UserSubscription) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"plan_id", "is_active", "updated_at"}),
		}).Create(sub).Error
	})
}

func (r *subscriptionRepository) AddUserConsumablePackage(ctx context.Context, userID, packageID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var pkg entities.ConsumablePackage
		if err := tx.First(&pkg, "id = ?", packageID).Error; err != nil {
			return err
		}

		cons := entities.UserConsumable{
			UserID:   userID,
			ItemType: pkg.ItemType,
			Amount:   pkg.Amount,
		}

		return tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}, {Name: "item_type"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"amount":     gorm.Expr("user_consumables.amount + ?", pkg.Amount),
				"updated_at": time.Now(),
			}),
		}).Create(&cons).Error
	})
}
