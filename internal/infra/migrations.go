package infra

import (
	"github.com/pipigendut/dating-backend/internal/entities"
	"gorm.io/gorm"
)

// AutoMigrate and apply optimized indexes
func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&entities.User{},
		&entities.Profile{},
		&entities.Photo{},
		&entities.AuthProvider{},
		&entities.Device{},
		&entities.RefreshToken{},
	)
	if err != nil {
		return err
	}

	// Manual index strategy for scalability
	// 1. Index on Email for login speed
	// 2. Composite index on Profile (Gender, DateOfBirth) for discovery queries
	// 3. Foreign key indexes (GORM does some, but being explicit is better at scale)

	db.Exec("CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_profiles_gender_dob ON profiles(gender, date_of_birth)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_photos_user_id_sort ON photos(user_id, sort_order)")

	return nil
}
