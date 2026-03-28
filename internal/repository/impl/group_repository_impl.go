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
	err := r.db.WithContext(ctx).
		Preload("Members.User.Photos").
		Preload("Members.User.Gender").
		Preload("Members.User.RelationshipType").
		Preload("Members.User.InterestedGenders").
		Preload("Members.User.Interests").
		Preload("Members.User.Languages").
		First(&group, "id = ?", id).Error
	return &group, err
}

func (r *groupRepository) GetGroupByEntityID(ctx context.Context, entityID uuid.UUID) (*entities.Group, error) {
	var group entities.Group
	err := r.db.WithContext(ctx).
		Preload("Members.User.Photos").
		Preload("Members.User.Gender").
		Preload("Members.User.RelationshipType").
		Preload("Members.User.InterestedGenders").
		Preload("Members.User.Interests").
		Preload("Members.User.Languages").
		First(&group, "entity_id = ?", entityID).Error
	return &group, err
}

func (r *groupRepository) GetGroupByUserID(ctx context.Context, userID uuid.UUID) (*entities.Group, error) {
	var group entities.Group
	err := r.db.WithContext(ctx).
		Table("groups").
		Select("groups.*").
		Joins("JOIN group_members gm ON gm.group_id = groups.id").
		Where("gm.user_id = ?", userID).
		Preload("Members.User.Photos").
		Preload("Members.User.Gender").
		Preload("Members.User.RelationshipType").
		Preload("Members.User.InterestedGenders").
		Preload("Members.User.Interests").
		Preload("Members.User.Languages").
		First(&group).Error
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
