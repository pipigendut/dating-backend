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
		// 1. Interactions (Swipes, Matches, Impressions, Unmatches)
		log.Printf("[UserCleanupJob] Stage 1: Deleting interactions for entity: %s", entityID)
		if entityID != uuid.Nil {
			if err := tx.Unscoped().Where("swiper_entity_id = ? OR swiped_entity_id = ?", entityID, entityID).Delete(&entities.Swipe{}).Error; err != nil {
				return fmt.Errorf("failed to delete swipes: %w", err)
			}
			if err := tx.Unscoped().Where("entity1_id = ? OR entity2_id = ?", entityID, entityID).Delete(&entities.Match{}).Error; err != nil {
				return fmt.Errorf("failed to delete matches: %w", err)
			}
			if err := tx.Unscoped().Where("viewer_entity_id = ? OR shown_entity_id = ?", entityID, entityID).Delete(&entities.EntityImpression{}).Error; err != nil {
				return fmt.Errorf("failed to delete impressions: %w", err)
			}
			if err := tx.Unscoped().Where("swiper_entity_id = ? OR target_entity_id = ?", entityID, entityID).Delete(&entities.EntityUnmatch{}).Error; err != nil {
				return fmt.Errorf("failed to delete unmatches: %w", err)
			}
		}

		// 2. Chat Data: delete user's conversations as participant
		log.Printf("[UserCleanupJob] Stage 2: Deleting chat data for user: %s", userID)
		var participants []entities.ConversationParticipant
		if err := tx.Where("user_id = ?", userID).Find(&participants).Error; err != nil {
			return fmt.Errorf("failed to find conversation participants: %w", err)
		}

		for _, cp := range participants {
			// Per-conversation cleanup
			// a. Delete last read record
			if err := tx.Unscoped().Where("conversation_id = ? AND user_id = ?", cp.ConversationID, userID).Delete(&entities.MessageRead{}).Error; err != nil {
				return fmt.Errorf("failed to delete message reads: %w", err)
			}
			
			// b. Identify conversation type
			var conv entities.Conversation
			if err := tx.Unscoped().First(&conv, "id = ?", cp.ConversationID).Error; err == nil {
				// If it's a direct chat, we purge the whole conversation
				if conv.Type == entities.ConversationTypeDirect {
					if err := tx.Unscoped().Where("conversation_id = ?", cp.ConversationID).Delete(&entities.Message{}).Error; err != nil {
						return fmt.Errorf("failed to delete messages in direct chat: %w", err)
					}
					if err := tx.Unscoped().Where("conversation_id = ?", cp.ConversationID).Delete(&entities.ConversationParticipant{}).Error; err != nil {
						return fmt.Errorf("failed to delete participants in direct chat: %w", err)
					}
					if err := tx.Unscoped().Delete(&conv).Error; err != nil {
						return fmt.Errorf("failed to delete direct conversation record: %w", err)
					}
				} else {
					// Group chat: only remove this user's trace
					if err := tx.Unscoped().Where("conversation_id = ? AND sender_id = ?", cp.ConversationID, userID).Delete(&entities.Message{}).Error; err != nil {
						return fmt.Errorf("failed to delete user messages in group: %w", err)
					}
					if err := tx.Unscoped().Delete(&cp).Error; err != nil {
						return fmt.Errorf("failed to remove user from group chat participants: %w", err)
					}
				}
			}
		}

		// 3. Disband Groups Created by User
		log.Printf("[UserCleanupJob] Stage 3: Disbanding groups created by user: %s", userID)
		var ownedGroups []entities.Group
		if err := tx.Where("created_by = ?", userID).Find(&ownedGroups).Error; err != nil {
			return fmt.Errorf("failed to find owned groups: %w", err)
		}
		for _, g := range ownedGroups {
			// 1. Delete Group Members
			if err := tx.Unscoped().Where("group_id = ?", g.ID).Delete(&entities.GroupMember{}).Error; err != nil {
				return fmt.Errorf("failed to delete group members: %w", err)
			}
			// 2. Delete Group Invites
			if err := tx.Unscoped().Where("group_id = ?", g.ID).Delete(&entities.GroupInvite{}).Error; err != nil {
				return fmt.Errorf("failed to delete group invites: %w", err)
			}
			
			// 3. Delete Group record itself
			if err := tx.Unscoped().Where("id = ?", g.ID).Delete(&entities.Group{}).Error; err != nil {
				return fmt.Errorf("failed to delete group record: %w", err)
			}
			
			groupEntityID := g.EntityID
			if groupEntityID != uuid.Nil {
				log.Printf("[UserCleanupJob] Stage 3.1: Deleting interactions and chat data for squad entity: %s", groupEntityID)
				
				// 4. Find all match IDs involving this entity
				var matchIDs []uuid.UUID
				if err := tx.Model(&entities.Match{}).
					Where("entity1_id = ? OR entity2_id = ?", groupEntityID, groupEntityID).
					Pluck("id", &matchIDs).Error; err != nil {
					return fmt.Errorf("failed to pluck match IDs for group: %w", err)
				}

				// Find all conversation IDs linked to this group (internal squad chat) or matches
				var convIDs []uuid.UUID
				if err := tx.Model(&entities.Conversation{}).Where("entity_id = ?", groupEntityID).Pluck("id", &convIDs).Error; err != nil {
					return fmt.Errorf("failed to pluck internal squad chat IDs: %w", err)
				}

				if len(matchIDs) > 0 {
					var matchConvIDs []uuid.UUID
					if err := tx.Model(&entities.Conversation{}).
						Where("entity_id IN ?", matchIDs).
						Pluck("id", &matchConvIDs).Error; err != nil {
						return fmt.Errorf("failed to pluck match conversation IDs for group: %w", err)
					}
					convIDs = append(convIDs, matchConvIDs...)
				}

				if len(convIDs) > 0 {
					// Delete message reads
					if err := tx.Unscoped().Where("conversation_id IN ?", convIDs).Delete(&entities.MessageRead{}).Error; err != nil {
						return fmt.Errorf("failed to delete message reads for group conversations: %w", err)
					}
					// 5. Delete participants
					if err := tx.Unscoped().Where("conversation_id IN ?", convIDs).Delete(&entities.ConversationParticipant{}).Error; err != nil {
						return fmt.Errorf("failed to delete participants for group conversations: %w", err)
					}
					// 6. Delete messages
					if err := tx.Unscoped().Where("conversation_id IN ?", convIDs).Delete(&entities.Message{}).Error; err != nil {
						return fmt.Errorf("failed to delete messages for group conversations: %w", err)
					}
					// 7. Delete conversations
					if err := tx.Unscoped().Where("id IN ?", convIDs).Delete(&entities.Conversation{}).Error; err != nil {
						return fmt.Errorf("failed to delete group conversations: %w", err)
					}
				}

				if len(matchIDs) > 0 {
					// 8. Delete Matches
					if err := tx.Unscoped().Where("entity1_id = ? OR entity2_id = ?", groupEntityID, groupEntityID).Delete(&entities.Match{}).Error; err != nil {
						return fmt.Errorf("failed to delete group matches: %w", err)
					}
				}

				// Delete other interactions
				if err := tx.Unscoped().Where("viewer_entity_id = ? OR shown_entity_id = ?", groupEntityID, groupEntityID).Delete(&entities.EntityImpression{}).Error; err != nil {
					return fmt.Errorf("failed to delete group impressions: %w", err)
				}
				if err := tx.Unscoped().Where("swiper_entity_id = ? OR target_entity_id = ?", groupEntityID, groupEntityID).Delete(&entities.EntityUnmatch{}).Error; err != nil {
					return fmt.Errorf("failed to delete group unmatches: %w", err)
				}
				
				// 9. Delete Swipes
				if err := tx.Unscoped().Where("swiper_entity_id = ? OR swiped_entity_id = ?", groupEntityID, groupEntityID).Delete(&entities.Swipe{}).Error; err != nil {
					return fmt.Errorf("failed to delete group swipes: %w", err)
				}

				// 10. Delete the Entity itself
				if err := tx.Unscoped().Where("id = ?", groupEntityID).Delete(&entities.Entity{}).Error; err != nil {
					return fmt.Errorf("failed to delete group entity: %w", err)
				}
			}
		}

		// 4. Remove User from Joined Groups & Invites
		log.Printf("[UserCleanupJob] Stage 4: Removing user from joined groups")
		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&entities.GroupMember{}).Error; err != nil {
			return fmt.Errorf("failed to remove user from groups: %w", err)
		}
		if err := tx.Unscoped().Where("inviter_id = ?", userID).Delete(&entities.GroupInvite{}).Error; err != nil {
			return fmt.Errorf("failed to delete user-created invites: %w", err)
		}

		// 5. Monetization & Usage Data
		log.Printf("[UserCleanupJob] Stage 5: Deleting monetization data")
		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&entities.UserSubscription{}).Error; err != nil {
			return fmt.Errorf("failed to delete subscriptions: %w", err)
		}
		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&entities.UserConsumable{}).Error; err != nil {
			return fmt.Errorf("failed to delete consumables: %w", err)
		}
		if entityID != uuid.Nil {
			if err := tx.Unscoped().Where("entity_id = ?", entityID).Delete(&entities.EntityBoost{}).Error; err != nil {
				return fmt.Errorf("failed to delete boosts: %w", err)
			}
		}

		// 6. Auth, Devices, Presence
		log.Printf("[UserCleanupJob] Stage 6: Deleting auth/device data")
		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&entities.AuthProvider{}).Error; err != nil {
			return fmt.Errorf("failed to delete auth providers: %w", err)
		}
		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&entities.Device{}).Error; err != nil {
			return fmt.Errorf("failed to delete devices: %w", err)
		}
		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&entities.RefreshToken{}).Error; err != nil {
			return fmt.Errorf("failed to delete refresh tokens: %w", err)
		}
		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&entities.UserPresence{}).Error; err != nil {
			return fmt.Errorf("failed to delete presence: %w", err)
		}

		// 7. Pivot tables
		log.Printf("[UserCleanupJob] Stage 7: Deleting pivot tables")
		if err := tx.Exec("DELETE FROM user_interests WHERE user_id = ?", userID).Error; err != nil {
			return fmt.Errorf("failed to delete interests: %w", err)
		}
		if err := tx.Exec("DELETE FROM user_interested_genders WHERE user_id = ?", userID).Error; err != nil {
			return fmt.Errorf("failed to delete interested genders: %w", err)
		}
		if err := tx.Exec("DELETE FROM user_languages WHERE user_id = ?", userID).Error; err != nil {
			return fmt.Errorf("failed to delete user languages: %w", err)
		}

		// 8. Photos and Storage
		log.Printf("[UserCleanupJob] Stage 8: Deleting photos and storage files")
		var photos []entities.Photo
		if err := tx.Where("user_id = ?", userID).Find(&photos).Error; err != nil {
			return fmt.Errorf("failed to find photos: %w", err)
		}
		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&entities.Photo{}).Error; err != nil {
			return fmt.Errorf("failed to delete photos from DB: %w", err)
		}
		
		ctxEx := context.Background()
		for _, photo := range photos {
			if idx := strings.Index(photo.URL, "users/"); idx != -1 {
				key := photo.URL[idx:]
				if h.storage != nil {
					_ = h.storage.DeleteFile(ctxEx, key)
				}
			}
		}

		log.Printf("[UserCleanupJob] Successfully wiped relations for user: %s", userID)
		return nil
	})
}
