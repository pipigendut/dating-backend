package impl

import (
	"context"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/repository"
	"gorm.io/gorm"
)

type groupRepository struct {
	db *gorm.DB
}

func NewGroupRepository(db *gorm.DB) repository.GroupRepository {
	return &groupRepository{db: db}
}

func (r *groupRepository) CreateGroup(ctx context.Context, group *entities.Group) error {
	return r.db.WithContext(ctx).Create(group).Error
}

func (r *groupRepository) GetGroupByID(ctx context.Context, id uuid.UUID) (*entities.Group, error) {
	var group entities.Group
	query := r.db.WithContext(ctx)
	query = ApplyFullUserPreload(query, "Members.User")
	err := query.First(&group, "id = ?", id).Error
	return &group, err
}

func (r *groupRepository) GetGroupByEntityID(ctx context.Context, entityID uuid.UUID) (*entities.Group, error) {
	var group entities.Group
	query := r.db.WithContext(ctx)
	query = ApplyFullUserPreload(query, "Members.User")
	err := query.First(&group, "entity_id = ?", entityID).Error
	return &group, err
}

func (r *groupRepository) GetGroupByUserID(ctx context.Context, userID uuid.UUID) (*entities.Group, error) {
	var group entities.Group
	query := r.db.WithContext(ctx).
		Table("groups").
		Select("groups.*").
		Joins("JOIN group_members gm ON gm.group_id = groups.id").
		Where("gm.user_id = ?", userID)
	
	query = ApplyFullUserPreload(query, "Members.User")
	err := query.First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *groupRepository) AddMember(ctx context.Context, member *entities.GroupMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *groupRepository) RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("group_id = ? AND user_id = ?", groupID, userID).Delete(&entities.GroupMember{}).Error
}

func (r *groupRepository) IsMember(ctx context.Context, groupID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entities.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *groupRepository) GetMembers(ctx context.Context, groupID uuid.UUID) ([]entities.GroupMember, error) {
	var members []entities.GroupMember
	err := r.db.WithContext(ctx).Preload("User.Photos").Where("group_id = ?", groupID).Find(&members).Error
	return members, err
}

func (r *groupRepository) CreateInvite(ctx context.Context, invite *entities.GroupInvite) error {
	return r.db.WithContext(ctx).Create(invite).Error
}

func (r *groupRepository) GetInviteByToken(ctx context.Context, token string) (*entities.GroupInvite, error) {
	var invite entities.GroupInvite
	err := r.db.WithContext(ctx).
		Preload("Group").
		Preload("Inviter").
		First(&invite, "token = ?", token).Error
	if err != nil {
		return nil, err
	}
	return &invite, nil
}

func (r *groupRepository) MarkInviteUsed(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).
		Model(&entities.GroupInvite{}).
		Where("token = ?", token).
		Update("used_at", gorm.Expr("NOW()")).Error
}

func (r *groupRepository) RemoveUserFromGroupConversations(ctx context.Context, groupEntityID uuid.UUID, userID uuid.UUID) error {
	// Find all conversations tied to this groupEntityID (for when the group entity itself was given a direct conv id)
	subQuery1 := r.db.Model(&entities.Conversation{}).Select("id").Where("entity_id = ?", groupEntityID)
	
	// Also delete from any conversation where this group matched someone else
	subQuery2 := r.db.Model(&entities.Conversation{}).
		Joins("JOIN matches m ON m.id = conversations.entity_id").
		Where("m.entity1_id = ? OR m.entity2_id = ?", groupEntityID, groupEntityID).
		Select("conversations.id")

	return r.db.WithContext(ctx).
		Where("user_id = ? AND (conversation_id IN (?) OR conversation_id IN (?))", userID, subQuery1, subQuery2).
		Delete(&entities.ConversationParticipant{}).Error
}

func (r *groupRepository) DisbandGroup(ctx context.Context, groupID uuid.UUID, entityID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Delete Group Members
		if err := tx.Where("group_id = ?", groupID).Delete(&entities.GroupMember{}).Error; err != nil {
			return err
		}

		// 2. Delete Group Invites
		if err := tx.Where("group_id = ?", groupID).Delete(&entities.GroupInvite{}).Error; err != nil {
			return err
		}

		// 3. Delete Group record itself
		if err := tx.Where("id = ?", groupID).Delete(&entities.Group{}).Error; err != nil {
			return err
		}

		// 4. Find all match IDs involving this entity
		var matchIDs []uuid.UUID
		if err := tx.Model(&entities.Match{}).
			Where("entity1_id = ? OR entity2_id = ?", entityID, entityID).
			Pluck("id", &matchIDs).Error; err != nil {
			return err
		}

		if len(matchIDs) > 0 {
			// Find all conversation IDs tied to those matches
			var convIDs []uuid.UUID
			if err := tx.Model(&entities.Conversation{}).
				Where("entity_id IN ?", matchIDs).
				Pluck("id", &convIDs).Error; err != nil {
				return err
			}

			if len(convIDs) > 0 {
				// 5. Delete participants first (FK constraint)
				if err := tx.Where("conversation_id IN ?", convIDs).
					Delete(&entities.ConversationParticipant{}).Error; err != nil {
					return err
				}

				// 6. Delete messages
				if err := tx.Where("conversation_id IN ?", convIDs).
					Delete(&entities.Message{}).Error; err != nil {
					return err
				}

				// 7. Delete conversations
				if err := tx.Where("id IN ?", convIDs).
					Delete(&entities.Conversation{}).Error; err != nil {
					return err
				}
			}

			// 8. Delete Matches
			if err := tx.Where("entity1_id = ? OR entity2_id = ?", entityID, entityID).
				Delete(&entities.Match{}).Error; err != nil {
				return err
			}
		}

		// 9. Delete Swipes
		if err := tx.Where("swiper_entity_id = ? OR swiped_entity_id = ?", entityID, entityID).
			Delete(&entities.Swipe{}).Error; err != nil {
			return err
		}

		// 10. Delete the Entity itself (last, due to FK dependencies)
		if err := tx.Where("id = ?", entityID).Delete(&entities.Entity{}).Error; err != nil {
			return err
		}

		return nil
	})
}
