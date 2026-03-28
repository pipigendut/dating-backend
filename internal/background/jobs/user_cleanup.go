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

	// Fetch user unscoped to get EntityID and Email
	var user entities.User
	if err := h.db.Unscoped().First(&user, "id = ?", userID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	entityID := user.EntityID

	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Delete matching interactions (using EntityID now)
		if entityID != uuid.Nil {
			if err := tx.Unscoped().Where("swiper_entity_id = ? OR swiped_entity_id = ?", entityID, entityID).Delete(&entities.Swipe{}).Error; err != nil {
				return fmt.Errorf("failed to delete swipes: %w", err)
			}
			if err := tx.Unscoped().Where("entity1_id = ? OR entity2_id = ?", entityID, entityID).Delete(&entities.Match{}).Error; err != nil {
				return fmt.Errorf("failed to delete matches: %w", err)
			}
			// Delete new entity interaction tables
			tx.Unscoped().Where("viewer_entity_id = ? OR shown_entity_id = ?", entityID, entityID).Delete(&entities.EntityImpression{})
			tx.Unscoped().Where("swiper_entity_id = ? OR target_entity_id = ?", entityID, entityID).Delete(&entities.EntityUnmatch{})
		}

		// 2. Chat Data: delete user's conversations as participant
		var convIDs []uuid.UUID
		tx.Model(&entities.ConversationParticipant{}).Where("user_id = ?", userID).Pluck("conversation_id", &convIDs)
		
		if len(convIDs) > 0 {
			tx.Unscoped().Where("conversation_id IN ?", convIDs).Delete(&entities.MessageRead{})
			tx.Unscoped().Where("conversation_id IN ?", convIDs).Delete(&entities.Message{})
			tx.Unscoped().Where("conversation_id IN ?", convIDs).Delete(&entities.ConversationParticipant{})
			tx.Unscoped().Where("id IN ?", convIDs).Delete(&entities.Conversation{})
		}

		// 3. Disband Groups Created by User
		var ownedGroups []entities.Group
		tx.Where("created_by = ?", userID).Find(&ownedGroups)
		for _, g := range ownedGroups {
			// Purge members and invites
			tx.Unscoped().Where("group_id = ?", g.ID).Delete(&entities.GroupMember{})
			tx.Unscoped().Where("group_id = ?", g.ID).Delete(&entities.GroupInvite{})
			// Purge Group's Entity
			if g.EntityID != uuid.Nil {
				tx.Unscoped().Where("id = ?", g.EntityID).Delete(&entities.Entity{})
			}
			// Purge Group itself
			tx.Unscoped().Where("id = ?", g.ID).Delete(&entities.Group{})
		}

		// 4. Remove User from Joined Groups & Invites
		tx.Unscoped().Where("user_id = ?", userID).Delete(&entities.GroupMember{})
		tx.Unscoped().Where("user_id = ?", userID).Delete(&entities.GroupInvite{})
		if user.Email != nil {
			tx.Unscoped().Where("target_email = ?", *user.Email).Delete(&entities.GroupInvite{})
		}

		// 5. Delete Monetization & Usage Data
		tx.Unscoped().Where("user_id = ?", userID).Delete(&entities.UserSubscription{})
		tx.Unscoped().Where("user_id = ?", userID).Delete(&entities.UserConsumable{})
		if entityID != uuid.Nil {
			tx.Unscoped().Where("entity_id = ?", entityID).Delete(&entities.EntityBoost{})
		}

		// 6. Delete Auth & Devices
		tx.Unscoped().Where("user_id = ?", userID).Delete(&entities.AuthProvider{})
		tx.Unscoped().Where("user_id = ?", userID).Delete(&entities.Device{})
		tx.Unscoped().Where("user_id = ?", userID).Delete(&entities.RefreshToken{})
		tx.Unscoped().Where("user_id = ?", userID).Delete(&entities.UserPresence{})

		// 7. Delete Pivot tables
		ctxEx := context.Background()
		tx.Exec("DELETE FROM user_interests WHERE user_id = ?", userID)
		tx.Exec("DELETE FROM user_interested_genders WHERE user_id = ?", userID)
		tx.Exec("DELETE FROM user_languages WHERE user_id = ?", userID)

		// 8. Delete Photos and S3 objects
		var photos []entities.Photo
		tx.Where("user_id = ?", userID).Find(&photos)
		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&entities.Photo{}).Error; err != nil {
			return fmt.Errorf("failed to delete photos from DB: %w", err)
		}
		for _, photo := range photos {
			if idx := strings.Index(photo.URL, "users/"); idx != -1 {
				key := photo.URL[idx:]
				if h.storage != nil {
					_ = h.storage.DeleteFile(ctxEx, key)
				}
			}
		}

		// 9. Purge User's personal Entity record
		if entityID != uuid.Nil {
			tx.Unscoped().Where("id = ?", entityID).Delete(&entities.Entity{})
		}

		// 10. Update User record to clear EntityID (marker for restoration)
		tx.Unscoped().Model(&entities.User{}).Where("id = ?", userID).Update("entity_id", nil)

		log.Printf("[UserCleanupJob] Successfully wiped relations for user: %s", userID)
		return nil
	})
}
