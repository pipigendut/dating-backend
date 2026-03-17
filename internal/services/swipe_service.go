package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
	"gorm.io/gorm"
)

type SwipeService interface {
	GetSwipeCandidates(ctx context.Context, userID uuid.UUID, limit int) ([]entities.User, error)
	CreateSwipe(ctx context.Context, swiperID, swipedID uuid.UUID, direction entities.SwipeDirection) (*entities.Match, error)
	GetIncomingLikes(ctx context.Context, userID uuid.UUID) ([]IncomingLike, error)
	UndoLastSwipe(ctx context.Context, userID uuid.UUID) (*entities.User, error)
	RecordImpressions(ctx context.Context, viewerID uuid.UUID, shownUserIDs []uuid.UUID) error
}

type swipeService struct {
	db     *gorm.DB
	config ConfigService
}

func NewSwipeService(db *gorm.DB, config ConfigService) SwipeService {
	return &swipeService{
		db:     db,
		config: config,
	}
}

// GetSwipeCandidates implements the Tinder-like weighted random matchmaking query
func (s *swipeService) GetSwipeCandidates(ctx context.Context, userID uuid.UUID, limit int) ([]entities.User, error) {
	// 1. Get configurations from fast in-memory cache
	premiumScore := s.config.GetInt("premium_score", 50)
	boostScore := s.config.GetInt("boost_score", 200)
	cdPremium := s.config.GetInt("cooldown_premium_minutes", 10)
	cdFree := s.config.GetInt("cooldown_free_minutes", 60)
	cdBoost := s.config.GetInt("cooldown_boost_minutes", 3)
	scoreWeight := s.config.GetFloat("score_weight", 0.7)
	randomWeight := s.config.GetFloat("random_weight", 0.3)

	var candidates []entities.User

	// The PostgreSQL Query
	query := `
		WITH recent_impressions AS (
			SELECT shown_user_id, MAX(shown_at) as last_shown
			FROM user_impressions
			WHERE viewer_id = ?
			GROUP BY shown_user_id
		),
		scored_users AS (
			SELECT 
				u.*,
				(
					-- Base Activity Score (decaying up to 7 days, max 100 points)
					GREATEST(0, 100 - EXTRACT(EPOCH FROM (NOW() - u.last_active_at))/3600) +
					-- Premium Score
					CASE WHEN u.is_premium = true THEN ? ELSE 0 END +
					-- Boost Score
					CASE WHEN u.boost_until > NOW() THEN ? ELSE 0 END
				) as raw_score,
				ri.last_shown
			FROM users u
			LEFT JOIN recent_impressions ri ON ri.shown_user_id = u.id
			WHERE u.id != ?
			AND u.status = 'active'
			-- Exclude already swiped
			AND NOT EXISTS (
				SELECT 1 FROM swipes s 
				WHERE s.swiper_id = ? AND s.swiped_id = u.id
			)
			-- Cooldown Rules: only show if last_shown is NULL or older than their specific cooldown
			AND (
				ri.last_shown IS NULL OR 
				NOW() - ri.last_shown > (
					CASE 
						WHEN u.boost_until > NOW() THEN CAST(? AS FLOAT) * INTERVAL '1 minute'
						WHEN u.is_premium = true THEN CAST(? AS FLOAT) * INTERVAL '1 minute'
						ELSE CAST(? AS FLOAT) * INTERVAL '1 minute'
					END
				)
			)
		)
		SELECT * FROM scored_users
		ORDER BY (raw_score * ?) + (RANDOM() * 100 * ?) DESC
		LIMIT ?
	`

	err := s.db.WithContext(ctx).Raw(
		query,
		userID,                   // recent_impressions viewer_id
		premiumScore, boostScore, // scores
		userID,   // u.id != ?
		userID,   // swipes.swiper_id = ?
		cdBoost,  // cooldowns
		cdPremium,
		cdFree,
		scoreWeight, // weights
		randomWeight,
		limit,
	).Scan(&candidates).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get swipe candidates: %w", err)
	}

	return candidates, nil
}

func (s *swipeService) CreateSwipe(ctx context.Context, swiperID, swipedID uuid.UUID, direction entities.SwipeDirection) (*entities.Match, error) {
	var match *entities.Match

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var swiper entities.User
		if err := tx.Select("is_premium", "boost_until", "last_active_at").Where("id = ?", swiperID).First(&swiper).Error; err != nil {
			return fmt.Errorf("swiper not found: %w", err)
		}

		// 1. Calculate Priority Score & Handle CRUSH limits
		priorityScore := 0
		if direction == entities.SwipeDirectionLike || direction == entities.SwipeDirectionCrush {
			// Premium Score
			if swiper.IsPremium {
				priorityScore += s.config.GetInt("premium_score", 50)
			}
			// Boost Score
			if swiper.BoostUntil != nil && swiper.BoostUntil.After(time.Now()) {
				priorityScore += s.config.GetInt("boost_score", 200)
			}

			if direction == entities.SwipeDirectionCrush {
				// Verify daily limit
				var todayCrushes int64
				startOfDay := time.Now().Truncate(24 * time.Hour) // Basic UTC start of day for check
				tx.Model(&entities.Swipe{}).
					Where("swiper_id = ? AND direction = ? AND created_at >= ?", swiperID, entities.SwipeDirectionCrush, startOfDay).
					Count(&todayCrushes)

				limit := s.config.GetInt("crush_limit_free", 1)
				if swiper.IsPremium {
					limit = s.config.GetInt("crush_limit_premium", 5)
				}

				if int(todayCrushes) >= limit {
					return fmt.Errorf("daily crush limit reached (limit: %d)", limit)
				}

				priorityScore += s.config.GetInt("crush_score_bonus", 500)
			}
		}

		// 2. Insert Swipe Record
		swipe := entities.Swipe{
			ID:            uuid.New(),
			SwiperID:      swiperID,
			SwipedID:      swipedID,
			Direction:     direction,
			PriorityScore: priorityScore,
		}
		if err := tx.Create(&swipe).Error; err != nil {
			return err
		}

		// 3. If LIKE or CRUSH, check for mutual match
		if direction == entities.SwipeDirectionLike || direction == entities.SwipeDirectionCrush {
			var reverseSwipe entities.Swipe
			err := tx.Where("swiper_id = ? AND swiped_id = ? AND direction IN ?", swipedID, swiperID, []entities.SwipeDirection{entities.SwipeDirectionLike, entities.SwipeDirectionCrush}).First(&reverseSwipe).Error
			
			if err == nil {
				// Mutual Like exists! Create a Match
				newMatch := entities.Match{
					ID:      uuid.New(),
					User1ID: swiperID,
					User2ID: swipedID, // Order could be normalized but keeping it simple
				}
				if err := tx.Create(&newMatch).Error; err != nil {
					return err
				}
				match = &newMatch
			} else if err != gorm.ErrRecordNotFound {
				return err
			}
		}

		return nil
	})

	return match, err
}

type IncomingLike struct {
	User          entities.User
	IsCrush       bool
	PriorityScore int
	CreatedAt     time.Time
}

func (s *swipeService) GetIncomingLikes(ctx context.Context, userID uuid.UUID) ([]IncomingLike, error) {
	var results []struct {
		entities.User
		RawDirection     string    `gorm:"column:direction"`
		RawPriorityScore int       `gorm:"column:priority_score"`
		RawCreatedAt     time.Time `gorm:"column:swipe_time"`
	}

	query := `
		SELECT 
			u.*,
			s.direction,
			s.priority_score,
			s.created_at as swipe_time
		FROM swipes s
		JOIN users u ON u.id = s.swiper_id
		WHERE s.swiped_id = ?
		AND s.direction IN ('LIKE', 'CRUSH')
		AND NOT EXISTS (
			-- Exclude if the current user already swiped on them (whether match or dislike)
			SELECT 1 FROM swipes my_swipe 
			WHERE my_swipe.swiper_id = ? AND my_swipe.swiped_id = s.swiper_id
		)
		ORDER BY s.priority_score DESC, s.created_at DESC
		LIMIT 100
	`

	err := s.db.WithContext(ctx).Raw(query, userID, userID).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get incoming likes: %w", err)
	}

	incomingLikes := make([]IncomingLike, len(results))
	for i, r := range results {
		incomingLikes[i] = IncomingLike{
			User:          r.User,
			IsCrush:       r.RawDirection == string(entities.SwipeDirectionCrush),
			PriorityScore: r.RawPriorityScore,
			CreatedAt:     r.RawCreatedAt,
		}
	}

	return incomingLikes, nil
}

func (s *swipeService) UndoLastSwipe(ctx context.Context, userID uuid.UUID) (*entities.User, error) {
	var undoneUser *entities.User

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Get the last created swipe
		var lastSwipe entities.Swipe
		err := tx.Where("swiper_id = ?", userID).Order("created_at DESC").First(&lastSwipe).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("no swipe history found to undo")
			}
			return err
		}

		// 2. Fetch user to check premium status & verify undo limits
		var swiper entities.User
		if err := tx.Select("is_premium").Where("id = ?", userID).First(&swiper).Error; err != nil {
			return err
		}

		var todayUndos int64
		startOfDay := time.Now().Truncate(24 * time.Hour)
		// Count how many swipes this user has SOFT DELETED today
		tx.Unscoped().Model(&entities.Swipe{}).
			Where("swiper_id = ? AND deleted_at >= ?", userID, startOfDay).
			Count(&todayUndos)

		limit := s.config.GetInt("undo_limit_free", 1)
		if swiper.IsPremium {
			limit = s.config.GetInt("undo_limit_premium", 10)
		}

		if int(todayUndos) >= limit {
			return fmt.Errorf("daily undo limit reached (limit: %d)", limit)
		}

		// 3. Delete the swipe (soft delete)
		if err := tx.Delete(&lastSwipe).Error; err != nil {
			return err
		}

		// 4. If it was a LIKE or CRUSH, check if a Match exists and delete it
		if lastSwipe.Direction == entities.SwipeDirectionLike || lastSwipe.Direction == entities.SwipeDirectionCrush {
			// Find match looking at permutations since it sorts LEAST/GREATEST but we query both just in case
			err := tx.Where("(user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)",
				lastSwipe.SwiperID, lastSwipe.SwipedID,
				lastSwipe.SwipedID, lastSwipe.SwiperID).
				Delete(&entities.Match{}).Error
			
			if err != nil && err != gorm.ErrRecordNotFound {
				return err
			}
		}

		// 5. Fetch the target user so the frontend knows who to put back in the deck
		var targetUser entities.User
		if err := tx.Preload("Photos").Where("id = ?", lastSwipe.SwipedID).First(&targetUser).Error; err != nil {
			return err
		}
		
		undoneUser = &targetUser
		return nil
	})

	return undoneUser, err
}

func (s *swipeService) RecordImpressions(ctx context.Context, viewerID uuid.UUID, shownUserIDs []uuid.UUID) error {
	if len(shownUserIDs) == 0 {
		return nil
	}

	impressions := make([]entities.UserImpression, 0, len(shownUserIDs))
	now := time.Now()
	for _, uid := range shownUserIDs {
		impressions = append(impressions, entities.UserImpression{
			ID:          uuid.New(),
			ViewerID:    viewerID,
			ShownUserID: uid,
			ShownAt:     now,
		})
	}

	// Batch insert impressions
	return s.db.WithContext(ctx).Create(&impressions).Error
}
