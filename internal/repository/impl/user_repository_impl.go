package impl

import (
	"context"
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
func (r *userRepo) GetWithRelations(id uuid.UUID) (*entities.User, error) {
	var user entities.User
	query := r.db.WithContext(context.Background())
	query = ApplyFullUserPreload(query, "")
	err := query.First(&user, "id = ?", id).Error
	return &user, err
}

func (r *userRepo) Update(user *entities.User) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Save base user fields, ignoring associations
		if err := tx.Omit("Photos", "InterestedGenders", "Interests", "Languages", "AuthProviders", "Devices", "RefreshTokens", "Gender", "RelationshipType", "Entity").Save(user).Error; err != nil {
			return err
		}

		// Sync Many2Many manually only if provided
		if user.InterestedGenders != nil {
			if err := tx.Model(user).Association("InterestedGenders").Replace(user.InterestedGenders); err != nil {
				return err
			}
		}
		if user.Interests != nil {
			if err := tx.Model(user).Association("Interests").Replace(user.Interests); err != nil {
				return err
			}
		}
		if user.Languages != nil {
			if err := tx.Model(user).Association("Languages").Replace(user.Languages); err != nil {
				return err
			}
		}

		// Sync Photos only if provided
		if user.Photos != nil {
			// 3. Reset is_main for all photos of this user to handle single-main-photo logic safely
			if err := tx.Model(&entities.Photo{}).Where("user_id = ?", user.ID).Update("is_main", false).Error; err != nil {
				return err
			}

			// 4. Handle Deletions: Get current IDs from request
			var photoIDs []uuid.UUID
			for _, p := range user.Photos {
				if p.ID != uuid.Nil {
					photoIDs = append(photoIDs, p.ID)
				}
			}

			// Delete photos stored in DB for this user but NOT in the current request list
			if len(photoIDs) > 0 {
				if err := tx.Unscoped().Where("user_id = ? AND id NOT IN ?", user.ID, photoIDs).Delete(&entities.Photo{}).Error; err != nil {
					return err
				}
			} else {
				if err := tx.Unscoped().Where("user_id = ?", user.ID).Delete(&entities.Photo{}).Error; err != nil {
					return err
				}
			}

			// 5. UPSERT photos from the request
			for i := range user.Photos {
				user.Photos[i].UserID = user.ID
				// Save performs UPSERT: Updates if ID exists, Creates otherwise
				if err := tx.Omit("CreatedAt").Save(&user.Photos[i]).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (r *userRepo) GetByEmail(email string) (*entities.User, error) {
	var user entities.User
	err := r.db.First(&user, "LOWER(email) = LOWER(?)", email).Error
	return &user, err
}

func (r *userRepo) GetByEmailUnscoped(email string) (*entities.User, error) {
	var user entities.User
	err := r.db.Unscoped().First(&user, "LOWER(email) = LOWER(?)", email).Error
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
		UserID:         userID,
		Provider:       provider,
		ProviderUserID: providerUserID,
	}
	return r.db.Create(&authProvider).Error
}

func (r *userRepo) CreateWithRelations(user *entities.User) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 0. Create the solo Entity first so that the FK on users(entity_id) is satisfied.
		//    Every new user gets their own Entity of type "user" (solo mode).
		soloEntity := entities.Entity{
			Type: entities.EntityTypeUser,
		}
		if err := tx.Create(&soloEntity).Error; err != nil {
			return err
		}
		user.EntityID = soloEntity.ID

		// 1. Create the user row (omit association tables handled below)
		if err := tx.Omit("Photos", "AuthProviders", "Entity").Create(user).Error; err != nil {
			return err
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


func (r *userRepo) UpdatePremiumStatus(id uuid.UUID, isPremium bool) error {
	return r.db.Model(&entities.User{}).Where("id = ?", id).Update("is_premium", isPremium).Error
}

func (r *userRepo) Delete(id uuid.UUID) error {
	user := entities.User{}
	user.ID = id
	return r.db.Select("Photos", "AuthProviders", "Devices", "RefreshTokens", "InterestedGenders", "Interests", "Languages").Delete(&user).Error
}

func (r *userRepo) DeletePhoto(photoID uuid.UUID) error {
	return r.db.Delete(&entities.Photo{}, "id = ?", photoID).Error
}

