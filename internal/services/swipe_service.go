package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SwipeService interface {
	GetSwipeCandidates(ctx context.Context, userID uuid.UUID, filter SwipeFilter, limit int) ([]entities.User, error)
	CreateSwipe(ctx context.Context, swiperID, swipedID uuid.UUID, direction entities.SwipeDirection) (*entities.Match, *entities.User, error)
	GetIncomingLikes(ctx context.Context, userID uuid.UUID, limit, offset int) ([]IncomingLike, error)
	GetLikesSent(ctx context.Context, userID uuid.UUID, limit, offset int) ([]SentLike, error)
	UnlikeUser(ctx context.Context, swiperID, swipedID uuid.UUID) error
	UndoLastSwipe(ctx context.Context, userID uuid.UUID) (*entities.User, error)
	RecordImpressions(ctx context.Context, viewerID uuid.UUID, shownUserIDs []uuid.UUID) error
	UnmatchUser(ctx context.Context, userID, targetUserID uuid.UUID) error
}

type SwipeFilter struct {
	Distance          *int
	MinAge            *int
	MaxAge            *int
	Genders           []uuid.UUID
	Interests         []uuid.UUID
	RelationshipTypes []uuid.UUID
	Latitude          *float64
	Longitude         *float64
	MinHeight         *int
	MaxHeight         *int
}

type SentLike struct {
	User      entities.User
	CreatedAt time.Time
	ExpiresAt time.Time
}

type swipeService struct {
	db           *gorm.DB
	config       ConfigService
	chat         ChatService
	subscription SubscriptionService
	swipeRepo    repository.SwipeRepository
}

func NewSwipeService(db *gorm.DB, config ConfigService, chat ChatService, subscription SubscriptionService, swipeRepo repository.SwipeRepository) SwipeService {
	return &swipeService{
		db:           db,
		config:       config,
		chat:         chat,
		subscription: subscription,
		swipeRepo:    swipeRepo,
	}
}

// GetSwipeCandidates implements the Tinder-like weighted random matchmaking query
func (s *swipeService) GetSwipeCandidates(ctx context.Context, userID uuid.UUID, filter SwipeFilter, limit int) ([]entities.User, error) {
	// 1. Get configurations from fast in-memory cache
	premiumScore := s.config.GetInt("premium_score", 50)
	boostScore := s.config.GetInt("boost_score", 200)
	cdPremium := s.config.GetInt("swipe_impression_cooldown_premium", 10)
	cdFree := s.config.GetInt("swipe_impression_cooldown_free", 60)
	cdBoost := s.config.GetInt("swipe_impression_cooldown_boost", 3)
	scoreWeight := s.config.GetFloat("score_weight", 0.7)
	randomWeight := s.config.GetFloat("random_weight", 0.3)
	dislikeRecycleMinutes := s.config.GetInt("dislike_recycle_minutes", 4320)

	// 1. Prepare dynamic filters (Genders)
	var searchGenders []uuid.UUID
	if len(filter.Genders) > 0 {
		searchGenders = filter.Genders
	} else {
		// Use user's interested genders as default
		if err := s.db.WithContext(ctx).Table("user_interested_genders").
			Where("user_id = ?", userID).
			Pluck("gender_id", &searchGenders).Error; err != nil {
			return nil, fmt.Errorf("failed to get user interests: %w", err)
		}
	}

	// Final fallback: if no filters and no profile preferences, show all active genders
	if len(searchGenders) == 0 {
		if err := s.db.WithContext(ctx).Table("master_genders").
			Where("is_active = true").
			Pluck("id", &searchGenders).Error; err != nil {
			return nil, fmt.Errorf("failed to get fallback genders: %w", err)
		}
	}

	// Default distance
	searchDistance := 50.0
	if filter.Distance != nil {
		searchDistance = float64(*filter.Distance)
	}

	var candidates []entities.User

	// 2. The PostgreSQL Query - Dynamically built to avoid NULL issues
	queryBase := `
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
	`
	args := []interface{}{
		userID, userID, premiumScore, boostScore, userID,
	}

	// Dynamic Filters
	whereClauses := ""
	if len(searchGenders) > 0 {
		whereClauses += " AND u.gender_id IN (?)"
		args = append(args, searchGenders)
	}

	if filter.MinAge != nil {
		whereClauses += " AND u.age >= ?"
		args = append(args, *filter.MinAge)
	}
	if filter.MaxAge != nil {
		whereClauses += " AND u.age <= ?"
		args = append(args, *filter.MaxAge)
	}
	if filter.MinHeight != nil {
		whereClauses += " AND u.height_cm >= ?"
		args = append(args, *filter.MinHeight)
	}
	if filter.MaxHeight != nil {
		whereClauses += " AND u.height_cm <= ?"
		args = append(args, *filter.MaxHeight)
	}

	if len(filter.Interests) > 0 {
		whereClauses += " AND EXISTS (SELECT 1 FROM user_interests ui WHERE ui.user_id = u.id AND ui.interest_id IN (?))"
		args = append(args, filter.Interests)
	}

	if len(filter.RelationshipTypes) > 0 {
		whereClauses += " AND u.relationship_type_id IN (?)"
		args = append(args, filter.RelationshipTypes)
	}

	// Distance Filter (Haversine)
	if filter.Latitude != nil && filter.Longitude != nil {
		whereClauses += ` AND (
			u.latitude IS NOT NULL AND u.longitude IS NOT NULL AND
			(6371 * acos(
				least(1.0, cos(radians(?)) * cos(radians(u.latitude)) * 
				cos(radians(u.longitude) - radians(?)) + 
				sin(radians(?)) * sin(radians(u.latitude)))
			)) <= ?
		)`
		args = append(args, *filter.Latitude, *filter.Longitude, *filter.Latitude, searchDistance)
	}

	// Exclude swiped
	whereClauses += ` AND NOT EXISTS (
		SELECT 1 FROM swipes s 
		WHERE s.swiper_id = ? AND s.swiped_id = u.id AND s.deleted_at IS NULL
		AND (
			s.direction IN ('LIKE', 'CRUSH')
			OR s.updated_at > NOW() - (CAST(? AS FLOAT) * INTERVAL '1 minute')
		)
	)`
	args = append(args, userID, float64(dislikeRecycleMinutes))

	// Cooldown Rules
	whereClauses += ` AND (
		ri.last_shown IS NULL OR 
		NOW() - ri.last_shown > (
			CASE 
				WHEN EXISTS (SELECT 1 FROM user_boosts ub WHERE ub.user_id = u.id AND ub.is_active = true AND ub.expired_at > NOW()) THEN CAST(? AS FLOAT) * INTERVAL '1 minute'
				WHEN u.is_premium = true THEN CAST(? AS FLOAT) * INTERVAL '1 minute'
				ELSE CAST(? AS FLOAT) * INTERVAL '1 minute'
			END
		)
	)`
	args = append(args, float64(cdBoost), float64(cdPremium), float64(cdFree))

	finalQuery := queryBase + whereClauses + `
		)
		SELECT * FROM scored_users
		ORDER BY (raw_score * ?) + (RANDOM() * 100 * ?) DESC
		LIMIT ?
	`
	args = append(args, scoreWeight, randomWeight, limit)

	err := s.db.WithContext(ctx).Raw(finalQuery, args...).Scan(&candidates).Error

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
		if err := tx.Select("is_premium", "last_active_at", "swipe_count_today", "updated_at").Where("id = ?", swiperID).First(&swiper).Error; err != nil {
			return fmt.Errorf("swiper not found: %w", err)
		}

		// 1. Calculate Ranking Score & Priority Score
		var rankingScore float64
		var priorityScore int

		if direction == entities.SwipeDirectionLike || direction == entities.SwipeDirectionCrush {
			// Base Activity Score (can be simplified here if we use the one from candidate search)
			rankingScore = 1.0
			if swiper.LastActiveAt.After(time.Now().Add(-24 * time.Hour)) {
				rankingScore += 1.0
			}

			// Priority Score based on Subscription/Crush
			hasPriority, _, _ := s.subscription.HasFeature(ctx, swiperID, "priority_likes")
			if hasPriority {
				priorityScore = 50
				rankingScore += float64(s.config.GetInt("premium_score", 50))
			}

			// Boost Score
			isBoosted, _ := s.subscription.IsBoosted(ctx, swiperID)
			if isBoosted {
				rankingScore += float64(s.config.GetInt("boost_score", 200))
			}

			if direction == entities.SwipeDirectionCrush {
				priorityScore = s.config.GetInt("crush_priority_score", 100)
				rankingScore += float64(s.config.GetInt("crush_score_bonus", 500))

				// Use consumable
				success, err := s.subscription.UseConsumable(ctx, swiperID, "crush")
				if err != nil {
					return err
				}
				if !success {
					return fmt.Errorf("No Crushes Remaining")
				}
			}
		}

		// 2. Anti-Cheat: Daily Swipe Limit
		hasUnlimited, _, _ := s.subscription.HasFeature(ctx, swiperID, "unlimited_likes")

		// Reset count if it's a new day
		currentCount := swiper.SwipeCountToday
		isNewDay := swiper.UpdatedAt.Before(time.Now().Truncate(24 * time.Hour))
		if isNewDay {
			currentCount = 0
		}

		if !hasUnlimited {
			limit := s.config.GetInt("max_limit_likes_free", 50)
			if currentCount >= limit {
				return fmt.Errorf("Daily Like Limit Reached")
			}
		}

		// Update swipe count
		var updateErr error
		if isNewDay {
			updateErr = tx.Model(&entities.User{}).Where("id = ?", swiperID).Updates(map[string]interface{}{
				"swipe_count_today": 1,
				"updated_at":        time.Now(),
			}).Error
		} else {
			updateErr = tx.Model(&entities.User{}).Where("id = ?", swiperID).Update("swipe_count_today", gorm.Expr("swipe_count_today + 1")).Error
		}

		if updateErr != nil {
			return updateErr
		}

		// 3. Insert or Update Swipe Record (Upsert to handle soft-deleted duplicates)
		swipe := entities.Swipe{
			SwiperID:      swiperID,
			SwipedID:      swipedID,
			Direction:     direction,
			RankingScore:  rankingScore,
			PriorityScore: priorityScore,
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "swiper_id"}, {Name: "swiped_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"direction":      direction,
				"ranking_score":  rankingScore,
				"priority_score": priorityScore,
				"deleted_at":     nil,
				"created_at":     time.Now(),
				"updated_at":     time.Now(),
			}),
		}).Create(&swipe).Error; err != nil {
			return err
		}

		// 3. If LIKE or CRUSH, check for mutual match
		if direction == entities.SwipeDirectionLike || direction == entities.SwipeDirectionCrush {
			var reverseSwipe entities.Swipe
			err := tx.Where("swiper_id = ? AND swiped_id = ? AND direction IN ?", swipedID, swiperID, []entities.SwipeDirection{entities.SwipeDirectionLike, entities.SwipeDirectionCrush}).First(&reverseSwipe).Error

			if err == nil {
				userLowID, userHighID := swiperID, swipedID
				if userLowID.String() > userHighID.String() {
					userLowID, userHighID = swipedID, swiperID
				}

				newMatch := entities.Match{
					UserLowID:  userLowID,
					UserHighID: userHighID,
					VisibleAt:  time.Now(),
				}
				if err := tx.Create(&newMatch).Error; err != nil {
					return err
				}
				match = &newMatch

				// 4. Auto-create Chat Conversation
				// We do this inside the transaction or immediately after.
				_, err = s.chat.GetOrCreateConversation(ctx, swiperID, swipedID, newMatch.VisibleAt)
				if err != nil {
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

func (s *swipeService) GetIncomingLikes(ctx context.Context, userID uuid.UUID, limit, offset int) ([]IncomingLike, error) {
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
		AND (
			s.updated_at < NOW() - (
				SELECT INTERVAL '1 minute' * (value->>0)::INT 
				FROM app_configs 
				WHERE key = (
					CASE 
						WHEN EXISTS (
							SELECT 1 FROM user_subscriptions us 
							JOIN subscription_plan_features spf ON us.plan_id = spf.plan_id 
							WHERE us.user_id = ? AND us.is_active = true AND us.expired_at > NOW() AND spf.feature_key = 'priority_likes'
						) THEN 'incoming_like_delay_premium' 
						ELSE 'incoming_like_delay_free' 
					END
				)
			)
		)
		AND NOT EXISTS (
			-- Exclude if the current user already swiped on them (whether match or dislike)
			SELECT 1 FROM swipes my_swipe 
			WHERE my_swipe.swiper_id = ? AND my_swipe.swiped_id = s.swiper_id
		)
		ORDER BY s.ranking_score DESC, s.created_at DESC
		LIMIT ? OFFSET ?
	`

	err := s.db.WithContext(ctx).Raw(query, userID, userID, userID, limit, offset).Scan(&results).Error
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
		// 1. Check for undo feature
		hasUndo, _, err := s.subscription.HasFeature(ctx, userID, "undo_swipe")
		if err != nil {
			return err
		}
		if !hasUndo {
			return fmt.Errorf("undo feature is restricted to premium users")
		}

		// 2. Get the last created swipe (including deleted ones to verify one-time undo)
		var lastOverall entities.Swipe
		err = tx.Unscoped().Where("swiper_id = ?", userID).Order("created_at DESC").First(&lastOverall).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("no swipe history found to undo")
			}
			return err
		}

		// If the most recent record is already deleted, it means the last action was an UNDO.
		if lastOverall.DeletedAt.Valid {
			return fmt.Errorf("only one undo allowed per swipe action. Swipe again to enable undo.")
		}

		lastSwipe := lastOverall

		// 3. Delete the swipe (hard delete to allow re-swipe)
		if err := tx.Unscoped().Delete(&lastSwipe).Error; err != nil {
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

func (s *swipeService) GetLikesSent(ctx context.Context, userID uuid.UUID, limit, offset int) ([]SentLike, error) {
	swipes, err := s.swipeRepo.GetLikesSent(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	if len(swipes) == 0 {
		return []SentLike{}, nil
	}

	expiryHours := s.config.GetInt("like_expiry_hours", 168)
	expiryDuration := time.Duration(expiryHours) * time.Hour

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
				ExpiresAt: sw.UpdatedAt.Add(expiryDuration),
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
	for _, uid := range shownUserIDs {
		impressions = append(impressions, entities.UserImpression{
			ViewerID:    viewerID,
			ShownUserID: uid,
		})
	}

	// Batch insert impressions with conflict resolution
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&impressions).Error
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
	conv, err := s.chat.GetOrCreateConversation(ctx, userID, targetUserID, time.Now())
	if err != nil {
		return fmt.Errorf("could not find conversation: %w", err)
	}

	// 3. Process unmatch in transaction
	return s.swipeRepo.UnmatchUser(ctx, userID, targetUserID, match.ID, conv.ID)
}
