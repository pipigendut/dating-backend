package repository

import (
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
)

type UserRepository interface {
	Create(user *entities.User) error
	GetByID(id uuid.UUID) (*entities.User, error)
	GetWithRelations(id uuid.UUID) (*entities.User, error)
	Update(user *entities.User) error
	GetByEmail(email string) (*entities.User, error)
	GetByProvider(provider, providerUserID string) (*entities.User, error)
	LinkProvider(userID uuid.UUID, provider, providerUserID string) error
	CreateWithRelations(user *entities.User) error
	Delete(id uuid.UUID) error
	DeletePhoto(photoID uuid.UUID) error
}
