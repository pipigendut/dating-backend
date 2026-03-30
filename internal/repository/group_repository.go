package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
)

type GroupRepository interface {
	CreateGroup(ctx context.Context, group *entities.Group) error
	GetGroupByID(ctx context.Context, id uuid.UUID) (*entities.Group, error)
	GetGroupByEntityID(ctx context.Context, entityID uuid.UUID) (*entities.Group, error)
	GetGroupByUserID(ctx context.Context, userID uuid.UUID) (*entities.Group, error)
	AddMember(ctx context.Context, member *entities.GroupMember) error
	RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error
	IsMember(ctx context.Context, groupID, userID uuid.UUID) (bool, error)
	GetMembers(ctx context.Context, groupID uuid.UUID) ([]entities.GroupMember, error)

	// Invite Methods
	CreateInvite(ctx context.Context, invite *entities.GroupInvite) error
	GetInviteByToken(ctx context.Context, token string) (*entities.GroupInvite, error)
	MarkInviteUsed(ctx context.Context, token string) error

	// Lifecycle Methods
	RemoveUserFromGroupConversations(ctx context.Context, groupEntityID uuid.UUID, userID uuid.UUID) error
	DisbandGroup(ctx context.Context, groupID uuid.UUID, entityID uuid.UUID) error
}
