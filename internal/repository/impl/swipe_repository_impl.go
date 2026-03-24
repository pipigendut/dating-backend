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

func (r *swipeRepository) GetMatch(ctx context.Context, userID, targetUserID uuid.UUID) (*entities.Match, error) {
	userLowID, userHighID := userID, targetUserID
	if userLowID.String() > userHighID.String() {
		userLowID, userHighID = targetUserID, userID
	}

	var match entities.Match
	err := r.db.WithContext(ctx).
		Where("user_low_id = ? AND user_high_id = ?", userLowID, userHighID).
		First(&match).Error
	if err != nil {
		return nil, err
	}
	return &match, nil
}

func (r *swipeRepository) GetUnmatch(ctx context.Context, userID, targetUserID uuid.UUID) (*entities.Unmatch, error) {
	var unmatch entities.Unmatch
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND target_user_id = ?", userID, targetUserID).
		First(&unmatch).Error
	if err != nil {
		return nil, err
	}
	return &unmatch, nil
}

func (r *swipeRepository) UnmatchUser(ctx context.Context, userID, targetUserID uuid.UUID, matchID uuid.UUID, conversationID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Soft delete match
		if err := tx.Delete(&entities.Match{}, matchID).Error; err != nil {
			return err
		}

		// 2. Soft delete conversation
		if err := tx.Delete(&entities.Conversation{}, conversationID).Error; err != nil {
			return err
		}

		// 3. Insert into unmatches table
		unmatch := entities.Unmatch{
			UserID:       userID,
			TargetUserID: targetUserID,
			MatchID:      matchID,
		}
		if err := tx.Create(&unmatch).Error; err != nil {
			return err
		}

		return nil
	})
}
func (r *swipeRepository) GetLikesSent(ctx context.Context, userID uuid.UUID, limit, offset int) ([]entities.Swipe, error) {
	var swipes []entities.Swipe
	// Find users current user liked but not yet matched
	err := r.db.WithContext(ctx).
		Where("swiper_id = ? AND direction IN ?", userID, []entities.SwipeDirection{entities.SwipeDirectionLike, entities.SwipeDirectionCrush}).
		Where("NOT EXISTS (SELECT 1 FROM matches m WHERE ((m.user_low_id = ? AND m.user_high_id = swipes.swiped_id) OR (m.user_low_id = swipes.swiped_id AND m.user_high_id = ?)))", userID, userID).
		Order("updated_at DESC").
		Limit(limit).Offset(offset).
		Find(&swipes).Error
	return swipes, err
}

func (r *swipeRepository) UnlikeUser(ctx context.Context, swiperID, swipedID uuid.UUID) error {
	// Only allowed if no match exists
	userLowID, userHighID := swiperID, swipedID
	if userLowID.String() > userHighID.String() {
		userLowID, userHighID = swipedID, swiperID
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var match entities.Match
		err := tx.Where("user_low_id = ? AND user_high_id = ?", userLowID, userHighID).First(&match).Error
		if err == nil {
			return gorm.ErrInvalidDB // Or custom error
		} else if err != gorm.ErrRecordNotFound {
			return err
		}

		// Soft delete the swipe
		return tx.Model(&entities.Swipe{}).Where("swiper_id = ? AND swiped_id = ?", swiperID, swipedID).Delete(&entities.Swipe{}).Error
	})
}
