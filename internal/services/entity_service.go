package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/repository"
)

type EntityService interface {
	CreateSoloEntity(ctx context.Context) (*entities.Entity, error)
	GetEntityByID(ctx context.Context, id uuid.UUID) (*entities.Entity, error)
}

type entityService struct {
	entityRepo repository.EntityRepository
	userRepo   repository.UserRepository
}

func NewEntityService(entityRepo repository.EntityRepository, userRepo repository.UserRepository) EntityService {
	return &entityService{
		entityRepo: entityRepo,
		userRepo:   userRepo,
	}
}

func (s *entityService) CreateSoloEntity(ctx context.Context) (*entities.Entity, error) {
	entity := &entities.Entity{
		Type: entities.EntityTypeUser,
	}
	if err := s.entityRepo.Create(ctx, entity); err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *entityService) GetEntityByID(ctx context.Context, id uuid.UUID) (*entities.Entity, error) {
	return s.entityRepo.GetByID(ctx, id)
}
