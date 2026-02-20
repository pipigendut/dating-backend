package impl

import (
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/repository"
	"gorm.io/gorm"
)

type userRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) repository.UserRepository {
	return &userRepo{db: db}
}

// GORM specific models to keep entity clean from tags
type userModel struct {
	ID           uuid.UUID `gorm:"primaryKey;type:uuid"`
	Email        *string   `gorm:"uniqueIndex"`
	PasswordHash *string
	Status       string
	CreatedAt    int64 `gorm:"autoCreateTime"`
	UpdatedAt    int64 `gorm:"autoUpdateTime"`
}

func (m userModel) TableName() string {
	return "users"
}

func (r *userRepo) Create(user *entities.User) error {
	// Map entity to model here if needed, or use tags if standard enough
	// For simplicity in this demo, we'll use the entity directly but show how to avoid N+1
	return r.db.Create(user).Error
}

func (r *userRepo) GetByID(id uuid.UUID) (*entities.User, error) {
	var user entities.User
	// No Preload here by default - avoid N+1 and slow queries
	err := r.db.First(&user, "id = ?", id).Error
	return &user, err
}

func (r *userRepo) GetWithProfile(id uuid.UUID) (*entities.User, error) {
	var user entities.User
	// Explicit Preload only when needed
	err := r.db.Preload("Profile").Preload("Photos").First(&user, "id = ?", id).Error
	return &user, err
}

func (r *userRepo) Update(user *entities.User) error {
	return r.db.Save(user).Error
}

func (r *userRepo) GetByEmail(email string) (*entities.User, error) {
	var user entities.User
	err := r.db.First(&user, "email = ?", email).Error
	return &user, err
}
