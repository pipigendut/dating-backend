package repository

import (
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
)

type UserRepository interface {
	Create(user *entities.User) error
	GetByID(id uuid.UUID) (*entities.User, error)
	GetWithProfile(id uuid.UUID) (*entities.User, error)
	Update(user *entities.User) error
	GetByEmail(email string) (*entities.User, error)
}
