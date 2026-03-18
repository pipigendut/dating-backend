package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/repository"
	"gorm.io/gorm"
)

type SwipeService interface {
	GetSwipeCandidates(ctx context.Context, userID uuid.UUID, limit int) ([]entities.User, error)
	CreateSwipe(ctx context.Context, swiperID, swipedID uuid.UUID, direction entities.SwipeDirection) (*entities.Match, *entities.User, error)
	GetIncomingLikes(ctx context.Context, userID uuid.UUID) ([]IncomingLike, error)
	GetLikesSent(ctx context.Context, userID uuid.UUID) ([]SentLike, error)
	UnlikeUser(ctx context.Context, swiperID, swipedID uuid.UUID) error
	UndoLastSwipe(ctx context.Context, userID uuid.UUID) (*entities.User, error)
	RecordImpressions(ctx context.Context, viewerID uuid.UUID, shownUserIDs []uuid.UUID) error
	UnmatchUser(ctx context.Context, userID, targetUserID uuid.UUID) error
}

type SentLike struct {
	User      entities.User
	CreatedAt time.Time
}

type swipeService struct {
	db        *gorm.DB
	config    ConfigService
	chat      ChatService
	swipeRepo repository.SwipeRepository
}

func NewSwipeService(db *gorm.DB, config ConfigService, chat ChatService, swipeRepo repository.SwipeRepository) SwipeService {
	return &swipeService{
		db:        db,
		config:    config,
		chat:      chat,
		swipeRepo: swipeRepo,
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

	// 1. Get user's gender interests
	var interestedGenderIDs []uuid.UUID
	if err := s.db.WithContext(ctx).Table("user_interested_genders").
		Where("user_id = ?", userID).
		Pluck("gender_id", &interestedGenderIDs).Error; err != nil {
		return nil, fmt.Errorf("failed to get user interests: %w", err)
	}

	// 1.1 Support "Everyone" by expanding the list
	// e0000000-0000-0000-0000-000000000003 is "everyone"
	everyoneID := uuid.MustParse("e0000000-0000-0000-0000-000000000003")
	isEveryone := false
	for _, id := range interestedGenderIDs {
		if id == everyoneID {
			isEveryone = true
			break
		}
	}

	if isEveryone {
		// If interested in everyone, just get all active genders
		var allGenderIDs []uuid.UUID
		if err := s.db.WithContext(ctx).Table("master_genders").
			Where("is_active = true").
			Pluck("id", &allGenderIDs).Error; err == nil {
			interestedGenderIDs = allGenderIDs
		}
	}

	// 2. The PostgreSQL Query
	query := `
		WITH recent_impressions AS (
			SELECT shown_user_id, MAX(shown_at) as last_shown
			FROM user_impressions
			WHERE viewer_id = ?
			GROUP BY shown_user_id
		),
		recent_unmatches AS (
			SELECT target_user_id, id FROM unmatches
			WHERE user_id = ?
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
					CASE WHEN EXISTS (SELECT 1 FROM user_boosts ub WHERE ub.user_id = u.id AND ub.is_active = true AND ub.expired_at > NOW()) THEN ? ELSE 0 END -
					-- Unmatch Penalty
					CASE WHEN ru.id IS NOT NULL THEN 1000 ELSE 0 END
				) as raw_score,
				ri.last_shown
			FROM users u
			LEFT JOIN recent_impressions ri ON ri.shown_user_id = u.id
			LEFT JOIN recent_unmatches ru ON ru.target_user_id = u.id
			WHERE u.id != ?
			AND u.status = 'active'
			AND u.gender_id IN (?)
			-- Exclude already swiped
			AND NOT EXISTS (
				SELECT 1 FROM swipes s 
				WHERE s.swiper_id = ? AND s.swiped_id = u.id AND s.deleted_at IS NULL
			)
			-- Cooldown Rules: only show if last_shown is NULL or older than their specific cooldown
			AND (
				ri.last_shown IS NULL OR 
				NOW() - ri.last_shown > (
					CASE 
						WHEN EXISTS (SELECT 1 FROM user_boosts ub WHERE ub.user_id = u.id AND ub.is_active = true AND ub.expired_at > NOW()) THEN CAST(? AS FLOAT) * INTERVAL '1 minute'
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
		userID,                   // recent_unmatches user_id
		premiumScore, boostScore, // scores
		userID,              // u.id != ?
		interestedGenderIDs, // u.gender_id IN (?)
		userID,              // swipes.swiper_id = ?
		cdBoost,             // cooldowns
		cdPremium,
		cdFree,
		scoreWeight, // weights
		randomWeight,
		limit,
	).Scan(&candidates).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get swipe candidates: %w", err)
	}

	if len(candidates) == 0 {
		return candidates, nil
	}

	var candidateIDs []uuid.UUID
	orderMap := make(map[uuid.UUID]int)
	for i, c := range candidates {
		candidateIDs = append(candidateIDs, c.ID)
		orderMap[c.ID] = i
	}

	var fullUsers []entities.User
	err = s.db.WithContext(ctx).
		Preload("Gender").
		Preload("RelationshipType").
		Preload("InterestedGenders").
		Preload("Interests").
		Preload("Languages").
		Preload("Photos").
		Where("id IN ?", candidateIDs).
		Find(&fullUsers).Error

	if err != nil {
		return nil, fmt.Errorf("failed to load candidates relations: %w", err)
	}

	// Restore original weighted-random order
	result := make([]entities.User, len(candidates))
	for _, u := range fullUsers {
		idx := orderMap[u.ID]
		// Retain the raw query injected fields (though not strictly needed right now)
		result[idx] = u
	}

	return result, nil
}

func (s *swipeService) CreateSwipe(ctx context.Context, swiperID, swipedID uuid.UUID, direction entities.SwipeDirection) (*entities.Match, *entities.User, error) {
	var match *entities.Match
	var matchedUser *entities.User

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var swiper entities.User
		if err := tx.Select("is_premium", "last_active_at").Where("id = ?", swiperID).First(&swiper).Error; err != nil {
			return fmt.Errorf("swiper not found: %w", err)
		}

		// 1. Calculate Ranking Score & Handle CRUSH limits
		var rankingScore float64
		if direction == entities.SwipeDirectionLike || direction == entities.SwipeDirectionCrush {
			// Premium Score
			if swiper.IsPremium {
				rankingScore += float64(s.config.GetInt("premium_score", 50))
			}
			// Boost Score
			var activeBoost int64
			tx.Model(&entities.UserBoost{}).
				Where("user_id = ? AND is_active = true AND expired_at > NOW()", swiperID).
				Count(&activeBoost)
			
			if activeBoost > 0 {
				rankingScore += float64(s.config.GetInt("boost_score", 200))
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

				rankingScore += float64(s.config.GetInt("crush_score_bonus", 500))
			}
		}

		// 2. Insert Swipe Record
		swipe := entities.Swipe{
			ID:           uuid.New(),
			SwiperID:     swiperID,
			SwipedID:     swipedID,
			Direction:    direction,
			RankingScore: rankingScore,
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
				// Enforce deterministic pair: lowID < highID
				userLowID, userHighID := swiperID, swipedID
				if userLowID.String() > userHighID.String() {
					userLowID, userHighID = swipedID, swiperID
				}

				newMatch := entities.Match{
					ID:         uuid.New(),
					UserLowID:  userLowID,
					UserHighID: userHighID,
				}
				if err := tx.Create(&newMatch).Error; err != nil {
					return err
				}
				match = &newMatch

				// 4. Auto-create Chat Conversation
				// We do this inside the transaction or immediately after.
				// For robustness, ensure conversation exists.
				_, err = s.chat.GetOrCreateConversation(ctx, swiperID, swipedID)
				if err != nil {
					// We might not want to fail the whole swipe if chat creation fails, 
					// but for now let's be strict.
					return fmt.Errorf("failed to create chat conversation: %w", err)
				}

				// Fetch matched user details for the response
				var u entities.User
				if err := tx.Preload("Photos").Where("id = ?", swipedID).First(&u).Error; err == nil {
					matchedUser = &u
				}
			} else if err != gorm.ErrRecordNotFound {
				return err
			}
		}

		return nil
	})

	return match, matchedUser, err
}

type IncomingLike struct {
	User         entities.User
	IsCrush      bool
	RankingScore float64
	CreatedAt    time.Time
}

func (s *swipeService) GetIncomingLikes(ctx context.Context, userID uuid.UUID) ([]IncomingLike, error) {
	var results []struct {
		entities.User
		RawDirection    string    `gorm:"column:direction"`
		RawRankingScore float64   `gorm:"column:ranking_score"`
		RawCreatedAt    time.Time `gorm:"column:swipe_time"`
	}

	query := `
		SELECT 
			u.*,
			s.direction,
			s.ranking_score,
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
		ORDER BY s.ranking_score DESC, s.created_at DESC
		LIMIT 100
	`

	err := s.db.WithContext(ctx).Raw(query, userID, userID).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get incoming likes: %w", err)
	}

	if len(results) == 0 {
		return []IncomingLike{}, nil
	}

	var userIDs []uuid.UUID
	orderMap := make(map[uuid.UUID]int)
	for i, r := range results {
		userIDs = append(userIDs, r.ID)
		orderMap[r.ID] = i
	}

	var fullUsers []entities.User
	err = s.db.WithContext(ctx).
		Preload("Gender").
		Preload("RelationshipType").
		Preload("InterestedGenders").
		Preload("Interests").
		Preload("Languages").
		Preload("Photos").
		Where("id IN ?", userIDs).
		Find(&fullUsers).Error

	if err != nil {
		return nil, fmt.Errorf("failed to load incoming likes relations: %w", err)
	}

	// Restore original ordered result
	incomingLikes := make([]IncomingLike, len(results))
	for _, u := range fullUsers {
		idx := orderMap[u.ID]
		incomingLikes[idx] = IncomingLike{
			User:         u,
			IsCrush:      results[idx].RawDirection == string(entities.SwipeDirectionCrush),
			RankingScore: results[idx].RawRankingScore,
			CreatedAt:    results[idx].RawCreatedAt,
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
			userLowID, userHighID := lastSwipe.SwiperID, lastSwipe.SwipedID
			if userLowID.String() > userHighID.String() {
				userLowID, userHighID = lastSwipe.SwipedID, lastSwipe.SwiperID
			}

			err := tx.Where("user_low_id = ? AND user_high_id = ?", userLowID, userHighID).
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

func (s *swipeService) GetLikesSent(ctx context.Context, userID uuid.UUID) ([]SentLike, error) {
	swipes, err := s.swipeRepo.GetLikesSent(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(swipes) == 0 {
		return []SentLike{}, nil
	}

	var userIDs []uuid.UUID
	for _, sw := range swipes {
		userIDs = append(userIDs, sw.SwipedID)
	}

	var users []entities.User
	err = s.db.WithContext(ctx).
		Preload("Photos").
		Preload("Gender").
		Where("id IN ?", userIDs).
		Find(&users).Error
	if err != nil {
		return nil, err
	}

	userMap := make(map[uuid.UUID]entities.User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	var results []SentLike
	for _, sw := range swipes {
		if u, ok := userMap[sw.SwipedID]; ok {
			results = append(results, SentLike{
				User:      u,
				CreatedAt: sw.CreatedAt,
			})
		}
	}

	return results, nil
}

func (s *swipeService) UnlikeUser(ctx context.Context, swiperID, swipedID uuid.UUID) error {
	return s.swipeRepo.UnlikeUser(ctx, swiperID, swipedID)
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

func (s *swipeService) UnmatchUser(ctx context.Context, userID, targetUserID uuid.UUID) error {
	// 1. Find the match
	match, err := s.swipeRepo.GetMatch(ctx, userID, targetUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("no match found between these users")
		}
		return err
	}

	// 2. Find the conversation (should exist if matched)
	conv, err := s.chat.GetOrCreateConversation(ctx, userID, targetUserID)
	if err != nil {
		return fmt.Errorf("could not find conversation: %w", err)
	}

	// 3. Process unmatch in transaction
	return s.swipeRepo.UnmatchUser(ctx, userID, targetUserID, match.ID, conv.ID)
}
