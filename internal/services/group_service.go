package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/repository"
)

type GroupService interface {
	CreateGroup(ctx context.Context, userID uuid.UUID, name string) (*entities.Group, error)
	InviteToGroup(ctx context.Context, groupID, userID uuid.UUID) error
	IsGroupMember(ctx context.Context, groupID, userID uuid.UUID) (bool, error)
	GetGroupByID(ctx context.Context, id uuid.UUID) (*entities.Group, error)
	GenerateInviteLink(ctx context.Context, groupID, inviterID uuid.UUID) (*entities.GroupInvite, error)
	ValidateInvite(ctx context.Context, token string) (*entities.GroupInvite, error)
	AcceptInvite(ctx context.Context, token string, userID uuid.UUID) error
	GetMyGroup(ctx context.Context, userID uuid.UUID) (*entities.Group, error)
}

type groupService struct {
	groupRepo  repository.GroupRepository
	entityRepo repository.EntityRepository
	userRepo   repository.UserRepository
}

func NewGroupService(groupRepo repository.GroupRepository, entityRepo repository.EntityRepository, userRepo repository.UserRepository) GroupService {
	return &groupService{
		groupRepo:  groupRepo,
		entityRepo: entityRepo,
		userRepo:   userRepo,
	}
}

func (s *groupService) CreateGroup(ctx context.Context, userID uuid.UUID, name string) (*entities.Group, error) {
	// 0. Check if user already has a group
	existing, _ := s.groupRepo.GetGroupByUserID(ctx, userID)
	if existing != nil {
		return nil, errors.New("user already belongs to a group")
	}

	// 1. Create Entity for group
	entity := &entities.Entity{
		Type: entities.EntityTypeGroup,
	}
	if err := s.entityRepo.Create(ctx, entity); err != nil {
		return nil, err
	}

	// 2. Create Group
	group := &entities.Group{
		EntityID:  entity.ID,
		Name:      name,
		CreatedBy: userID,
	}
	if err := s.groupRepo.CreateGroup(ctx, group); err != nil {
		return nil, err
	}

	// 3. Add creator as member and admin
	member := &entities.GroupMember{
		GroupID: group.ID,
		UserID:  userID,
		IsAdmin: true,
	}
	if err := s.groupRepo.AddMember(ctx, member); err != nil {
		return nil, err
	}

	return group, nil
}

func (s *groupService) InviteToGroup(ctx context.Context, groupID, userID uuid.UUID) error {
	member := &entities.GroupMember{
		GroupID: groupID,
		UserID:  userID,
		IsAdmin: false,
	}
	return s.groupRepo.AddMember(ctx, member)
}

func (s *groupService) IsGroupMember(ctx context.Context, groupID, userID uuid.UUID) (bool, error) {
	return s.groupRepo.IsMember(ctx, groupID, userID)
}

func (s *groupService) GetGroupByID(ctx context.Context, id uuid.UUID) (*entities.Group, error) {
	return s.groupRepo.GetGroupByID(ctx, id)
}

func (s *groupService) GenerateInviteLink(ctx context.Context, groupID, inviterID uuid.UUID) (*entities.GroupInvite, error) {
	// Generate secure token
	token := uuid.New().String()

	invite := &entities.GroupInvite{
		GroupID:   groupID,
		InviterID: inviterID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hours expiry
	}

	if err := s.groupRepo.CreateInvite(ctx, invite); err != nil {
		return nil, err
	}

	return invite, nil
}

func (s *groupService) ValidateInvite(ctx context.Context, token string) (*entities.GroupInvite, error) {
	invite, err := s.groupRepo.GetInviteByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	if invite.UsedAt != nil {
		return nil, errors.New("invite already used")
	}

	if time.Now().After(invite.ExpiresAt) {
		return nil, errors.New("invite expired")
	}

	return invite, nil
}

func (s *groupService) AcceptInvite(ctx context.Context, token string, userID uuid.UUID) error {
	invite, err := s.ValidateInvite(ctx, token)
	if err != nil {
		return err
	}

	// Check if already member
	isMember, err := s.groupRepo.IsMember(ctx, invite.GroupID, userID)
	if err != nil {
		return err
	}
	if isMember {
		return errors.New("already a member of this group")
	}

	// Add member
	member := &entities.GroupMember{
		GroupID: invite.GroupID,
		UserID:  userID,
		IsAdmin: false,
	}
	if err := s.groupRepo.AddMember(ctx, member); err != nil {
		return err
	}

	// Mark used
	return s.groupRepo.MarkInviteUsed(ctx, token)
}

func (s *groupService) GetMyGroup(ctx context.Context, userID uuid.UUID) (*entities.Group, error) {
	group, err := s.groupRepo.GetGroupByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return group, nil
}
