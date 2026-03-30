package impl

import (
	"context"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/repository"
	"gorm.io/gorm"
)

type swipeRepository struct {
	db *gorm.DB
}

func NewSwipeRepository(db *gorm.DB) repository.SwipeRepository {
	return &swipeRepository{db: db}
}

func (r *swipeRepository) CreateSwipe(ctx context.Context, swipe *entities.Swipe) error {
	return r.db.WithContext(ctx).Create(swipe).Error
}

func (r *swipeRepository) GetMatch(ctx context.Context, entity1ID, entity2ID uuid.UUID) (*entities.Match, error) {
	// Use deterministic ordering for unique indexes
	id1, id2 := entity1ID, entity2ID
	if id1.String() > id2.String() {
		id1, id2 = entity2ID, entity1ID
	}

	var match entities.Match
	err := r.db.WithContext(ctx).
		Where("entity1_id = ? AND entity2_id = ?", id1, id2).
		First(&match).Error
	if err != nil {
		return nil, err
	}
	return &match, nil
}

func (r *swipeRepository) GetLikesSent(ctx context.Context, entityIDs []uuid.UUID, limit, offset, expiryHours int) ([]entities.Swipe, error) {
	var swipes []entities.Swipe
	err := r.db.WithContext(ctx).
		Debug().
		Where("swiper_entity_id IN ? AND direction IN ?", entityIDs, []entities.SwipeDirection{entities.SwipeDirectionLike, entities.SwipeDirectionCrush}).
		Where("updated_at > NOW() - CAST(? AS FLOAT) * INTERVAL '1 hour'", expiryHours).
		Where("NOT EXISTS (SELECT 1 FROM matches m WHERE (m.entity1_id = swipes.swiper_entity_id AND m.entity2_id = swipes.swiped_entity_id) OR (m.entity1_id = swipes.swiped_entity_id AND m.entity2_id = swipes.swiper_entity_id))").
		Order("updated_at DESC").
		Limit(limit).Offset(offset).
		Find(&swipes).Error
	return swipes, err
}

func (r *swipeRepository) GetLikesYou(ctx context.Context, entityIDs []uuid.UUID, limit, offset, expiryHours int) ([]entities.Swipe, error) {
	var swipes []entities.Swipe
	err := r.db.WithContext(ctx).
		Where("swiped_entity_id IN ? AND direction IN ?", entityIDs, []entities.SwipeDirection{entities.SwipeDirectionLike, entities.SwipeDirectionCrush}).
		Where("updated_at > NOW() - CAST(? AS FLOAT) * INTERVAL '1 hour'", expiryHours).
		Where("NOT EXISTS (SELECT 1 FROM matches m WHERE (m.entity1_id = swipes.swiper_entity_id AND m.entity2_id = swipes.swiped_entity_id) OR (m.entity1_id = swipes.swiped_entity_id AND m.entity2_id = swipes.swiper_entity_id))").
		Order("updated_at DESC").
		Limit(limit).Offset(offset).
		Find(&swipes).Error
	return swipes, err
}

func (r *swipeRepository) CountLikesYou(ctx context.Context, entityIDs []uuid.UUID, expiryHours int) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.Swipe{}).
		Where("swiped_entity_id IN ? AND direction IN ?", entityIDs, []entities.SwipeDirection{entities.SwipeDirectionLike, entities.SwipeDirectionCrush}).
		Where("created_at > NOW() - CAST(? AS FLOAT) * INTERVAL '1 hour'", expiryHours).
		Where("NOT EXISTS (SELECT 1 FROM matches m WHERE (m.entity1_id = swipes.swiper_entity_id AND m.entity2_id = swipes.swiped_entity_id) OR (m.entity1_id = swipes.swiped_entity_id AND m.entity2_id = swipes.swiper_entity_id))").
		Count(&count).Error
	return count, err
}

func (r *swipeRepository) DeleteMatch(ctx context.Context, entity1ID, entity2ID uuid.UUID) error {
	id1, id2 := entity1ID, entity2ID
	if id1.String() > id2.String() {
		id1, id2 = entity2ID, entity1ID
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Find Match to get ID
		var match entities.Match
		if err := tx.Where("entity1_id = ? AND entity2_id = ?", id1, id2).First(&match).Error; err == nil {
			// 2. Clear Conversations associated with this Match
			// First delete participants and messages to avoid orphaned records
			var conv entities.Conversation
			if err := tx.Where("entity_id = ?", match.ID).First(&conv).Error; err == nil {
				tx.Unscoped().Where("conversation_id = ?", conv.ID).Delete(&entities.ConversationParticipant{})
				tx.Unscoped().Where("conversation_id = ?", conv.ID).Delete(&entities.Message{})
				tx.Unscoped().Delete(&conv)
			}

			// 3. Delete Match
			if err := tx.Unscoped().Delete(&match).Error; err != nil {
				return err
			}
		}

		// 4. Clear swipes
		if err := tx.Where("(swiper_entity_id = ? AND swiped_entity_id = ?) OR (swiper_entity_id = ? AND swiped_entity_id = ?)", id1, id2, id2, id1).Delete(&entities.Swipe{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *swipeRepository) DeleteSwipe(ctx context.Context, swiperEntityID, targetEntityID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("swiper_entity_id = ? AND swiped_entity_id = ?", swiperEntityID, targetEntityID).
		Delete(&entities.Swipe{}).Error
}
