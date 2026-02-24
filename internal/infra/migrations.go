package infra

import (
	"github.com/pipigendut/dating-backend/internal/entities"
	"gorm.io/gorm"
)

// AutoMigrate and apply optimized indexes
func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&entities.User{},
		&entities.Photo{},
		&entities.AuthProvider{},
		&entities.Device{},
		&entities.RefreshToken{},

		// Master Tables
		&entities.MasterGender{},
		&entities.MasterRelationshipType{},
		&entities.MasterInterest{},
		&entities.MasterLanguage{},

		// Pivot Tables
		&entities.UserInterestedGender{},
		&entities.UserInterest{},
		&entities.UserLanguage{},

		// Matching System
		&entities.Swipe{},
		&entities.Match{},
	)
	if err != nil {
		return err
	}

	// Manual index strategy for scalability
	// 1. Index on Email for login speed
	// 2. Composite indexes for matching and discovery queries

	db.Exec("CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_photos_user_id_sort ON photos(user_id, sort_order)")

	// ERD Optimizations
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_swipe ON swipes(swiper_id, swiped_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_swipe_received ON swipes(swiped_id, direction)")
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_match ON matches(LEAST(user1_id, user2_id), GREATEST(user1_id, user2_id))")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_user_discovery ON users(status, gender_id, relationship_type_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_photos_user_id_sort ON photos(user_id, sort_order)")

	return nil
}
