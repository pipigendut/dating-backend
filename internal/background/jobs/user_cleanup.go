package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/pipigendut/dating-backend/internal/background"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/infra/storage"
	"gorm.io/gorm"
)

const TaskUserCleanup = "user:cleanup"

type UserCleanupPayload struct {
	background.BaseJobPayload
	UserID uuid.UUID `json:"user_id"`
}

type UserCleanupHandler struct {
	db      *gorm.DB
	storage storage.StorageProvider
}

func NewUserCleanupHandler(db *gorm.DB, storage storage.StorageProvider) *UserCleanupHandler {
	return &UserCleanupHandler{
		db:      db,
		storage: storage,
	}
}

func (h *UserCleanupHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var p UserCleanupPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	userID := p.UserID
	log.Printf("[UserCleanupJob] Reclaiming data for user: %s", userID)

	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Delete matching interactions
		if err := tx.Unscoped().Where("swiper_id = ? OR swiped_id = ?", userID, userID).Delete(&entities.Swipe{}).Error; err != nil {
			return fmt.Errorf("failed to delete swipes: %w", err)
		}
		if err := tx.Unscoped().Where("user_low_id = ? OR user_high_id = ?", userID, userID).Delete(&entities.Match{}).Error; err != nil {
			return fmt.Errorf("failed to delete matches: %w", err)
		}
		if err := tx.Unscoped().Where("user_id = ? OR target_user_id = ?", userID, userID).Delete(&entities.Unmatch{}).Error; err != nil {
			return fmt.Errorf("failed to delete unmatches: %w", err)
		}
		if err := tx.Unscoped().Where("viewer_id = ? OR shown_user_id = ?", userID, userID).Delete(&entities.UserImpression{}).Error; err != nil {
			return fmt.Errorf("failed to delete impressions: %w", err)
		}

		// 2. Chat Data: we delete conversations where the user is a participant. 
		// If we delete the conversation, we must cascade to its messages and reads.
		var convIDs []uuid.UUID
		tx.Model(&entities.ConversationParticipant{}).Where("user_id = ?", userID).Pluck("conversation_id", &convIDs)
		
		if len(convIDs) > 0 {
			// Delete reads
			tx.Unscoped().Where("conversation_id IN ?", convIDs).Delete(&entities.MessageRead{})
			// Delete messages
			tx.Unscoped().Where("conversation_id IN ?", convIDs).Delete(&entities.Message{})
			// Delete participants
			tx.Unscoped().Where("conversation_id IN ?", convIDs).Delete(&entities.ConversationParticipant{})
			// Delete conversation
			tx.Unscoped().Where("id IN ?", convIDs).Delete(&entities.Conversation{})
		}

		// 3. Delete Monetization & Usage Data
		tx.Unscoped().Where("user_id = ?", userID).Delete(&entities.UserSubscription{})
		tx.Unscoped().Where("user_id = ?", userID).Delete(&entities.UserConsumable{})
		tx.Unscoped().Where("user_id = ?", userID).Delete(&entities.UserBoost{})

		// 4. Delete Auth & Devices
		tx.Unscoped().Where("user_id = ?", userID).Delete(&entities.AuthProvider{})
		tx.Unscoped().Where("user_id = ?", userID).Delete(&entities.Device{})
		tx.Unscoped().Where("user_id = ?", userID).Delete(&entities.RefreshToken{})
		tx.Unscoped().Where("user_id = ?", userID).Delete(&entities.UserPresence{})

		// 5. Delete Pivot tables (manually bypassing GORM relations for speed)
		tx.Exec("DELETE FROM user_interests WHERE user_id = ?", userID)
		tx.Exec("DELETE FROM user_interested_genders WHERE user_id = ?", userID)
		tx.Exec("DELETE FROM user_languages WHERE user_id = ?", userID)

		// 6. Delete Photos and S3 objects
		var photos []entities.Photo
		tx.Where("user_id = ?", userID).Find(&photos)

		// Delete photo records
		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&entities.Photo{}).Error; err != nil {
			return fmt.Errorf("failed to delete photos from DB: %w", err)
		}

		// Delete from S3/Oracle Object Storage
		for _, photo := range photos {
			// Extract key from URL. Example: public_bucket.com/users/uuid/profile/123.jpg => users/uuid/...
			// This is a naive split based on "users/". Needs to match GetUploadURL generation pattern.
			if idx := strings.Index(photo.URL, "users/"); idx != -1 {
				key := photo.URL[idx:]
				// Best effort delete
				if h.storage != nil {
					_ = h.storage.DeleteFile(ctx, key)
				}
			}
		}

		log.Printf("[UserCleanupJob] Successfully wiped relations for user: %s", userID)
		return nil
	})
}
