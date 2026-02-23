package impl

import (
	"time"

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
	// FullSaveAssociations ensures that Profile and Photos are also updated
	return r.db.Session(&gorm.Session{FullSaveAssociations: true}).Save(user).Error
}

func (r *userRepo) GetByEmail(email string) (*entities.User, error) {
	var user entities.User
	err := r.db.First(&user, "email = ?", email).Error
	return &user, err
}

func (r *userRepo) GetByProvider(provider, providerUserID string) (*entities.User, error) {
	var authProvider entities.AuthProvider
	err := r.db.Where("provider = ? AND provider_user_id = ?", provider, providerUserID).First(&authProvider).Error
	if err != nil {
		return nil, err
	}

	return r.GetByID(authProvider.UserID)
}

func (r *userRepo) LinkProvider(userID uuid.UUID, provider, providerUserID string) error {
	authProvider := entities.AuthProvider{
		ID:             uuid.New(),
		UserID:         userID,
		Provider:       provider,
		ProviderUserID: providerUserID,
		CreatedAt:      time.Now(),
	}
	return r.db.Create(&authProvider).Error
}

func (r *userRepo) CreateWithProfile(user *entities.User) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Omit associations to handle them manually below and avoid GORM's "smart" upsert issues
		if err := tx.Omit("Profile", "Photos", "AuthProviders").Create(user).Error; err != nil {
			return err
		}

		if user.Profile != nil {
			user.Profile.UserID = user.ID
			if err := tx.Create(user.Profile).Error; err != nil {
				return err
			}
		}

		if len(user.Photos) > 0 {
			for i := range user.Photos {
				user.Photos[i].UserID = user.ID
				if err := tx.Create(&user.Photos[i]).Error; err != nil {
					return err
				}
			}
		}

		if len(user.AuthProviders) > 0 {
			for i := range user.AuthProviders {
				user.AuthProviders[i].UserID = user.ID
				if err := tx.Create(&user.AuthProviders[i]).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (r *userRepo) Delete(id uuid.UUID) error {
	return r.db.Delete(&entities.User{}, "id = ?", id).Error
}
