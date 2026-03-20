package infra

import (
	"github.com/pipigendut/dating-backend/internal/entities"
	"gorm.io/gorm"
)

// AutoMigrate and apply optimized indexes
func Migrate(db *gorm.DB) error {
	// Manual pre-migration steps for refactored entities
	// Drop legacy primary keys that are not 'id' (from BaseModel)
	tablesToFix := map[string]string{
		"app_configs":                "key",
		"master_genders":             "code",
		"master_relationship_types": "code",
		"master_languages":          "code",
		"master_interests":          "name",
	}

	for table, legacyPK := range tablesToFix {
		// Drop the PK constraint if it's named like 'table_pkey' or if legacyPK is part of it
		db.Exec(`
			DO $$ 
			BEGIN 
				IF EXISTS (
					SELECT 1 
					FROM information_schema.table_constraints tc
					JOIN information_schema.key_column_usage kcu 
					  ON tc.constraint_name = kcu.constraint_name 
					 AND tc.table_name = kcu.table_name
					WHERE tc.table_name = '` + table + `' 
					  AND tc.constraint_type = 'PRIMARY KEY'
					  AND kcu.column_name = '` + legacyPK + `'
				) THEN
					ALTER TABLE ` + table + ` DROP CONSTRAINT IF EXISTS ` + table + `_pkey CASCADE;
				END IF;
			END $$;
		`)
	}

	// 3. Subscription & Consumable Refactor
	// Truncate to prevent unique constraint failures when applying the new idx_user_subs and idx_user_cons
	db.Exec("TRUNCATE TABLE user_subscriptions CASCADE")
	db.Exec("TRUNCATE TABLE user_consumables CASCADE")
	db.Exec("DROP TABLE IF EXISTS consumable_items CASCADE")

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
		&entities.UserImpression{},
		&entities.UserBoost{},
		&entities.AppConfig{},
		&entities.Unmatch{},

		// Monetization
		&entities.SubscriptionPlan{},
		&entities.SubscriptionPlanFeature{},
		&entities.SubscriptionPrice{},
		&entities.UserSubscription{},
		&entities.UserConsumable{},
		&entities.ConsumablePackage{},

		// Chat System
		&entities.Conversation{},
		&entities.ConversationParticipant{},
		&entities.Message{},
		&entities.MessageRead{},
		&entities.UserPresence{},
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
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_swipe ON swipes(swiper_id, swiped_id) WHERE deleted_at IS NULL")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_swipe_received ON swipes(swiped_id, direction) WHERE deleted_at IS NULL")
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_user_pair ON matches(user_low_id, user_high_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_user_discovery ON users(status, gender_id, age)")

	// Update age for existing users if still 0
	db.Exec("UPDATE users SET age = EXTRACT(YEAR FROM AGE(date_of_birth)) WHERE age = 0 AND date_of_birth IS NOT NULL")

	// Chat System Optimizations
	db.Exec("CREATE INDEX IF NOT EXISTS idx_messages_conversation_id_created ON messages(conversation_id, created_at DESC)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_participants_user_id ON conversation_participants(user_id)")
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_participant ON conversation_participants(conversation_id, user_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_message_reads_msg_user ON message_reads(message_id, user_id)")
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_unmatch ON unmatches(user_id, target_user_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_unmatch_match_id ON unmatches(match_id)")

	// Session System Optimizations
	db.Exec("DROP INDEX IF EXISTS idx_devices_device_id")
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_device_user ON devices(device_id, user_id)")

	// Cleanup legacy columns
	db.Exec("ALTER TABLE subscription_plan_features DROP COLUMN IF EXISTS value")
	db.Exec("ALTER TABLE user_consumables DROP COLUMN IF EXISTS type")
	db.Exec("ALTER TABLE user_consumables DROP COLUMN IF EXISTS remaining")
	db.Exec("ALTER TABLE user_consumables DROP COLUMN IF EXISTS expired_at")

	// Cleanup legacy columns in the renamed master table
	db.Exec("ALTER TABLE consumable_packages DROP COLUMN IF EXISTS type")
	db.Exec("ALTER TABLE consumable_packages DROP COLUMN IF EXISTS quantity")
	db.Exec("ALTER TABLE consumable_packages DROP COLUMN IF EXISTS remaining")
	db.Exec("ALTER TABLE consumable_packages DROP COLUMN IF EXISTS expired_at")
	db.Exec("ALTER TABLE consumable_packages DROP COLUMN IF EXISTS currency")
	db.Exec("ALTER TABLE consumable_packages DROP COLUMN IF EXISTS external_slug")
	db.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS verified_at TIMESTAMP WITH TIME ZONE")

	// Cleanup illegal genders (Everyone is not a gender, Other is being removed as per request)
	db.Exec("DELETE FROM master_genders WHERE code NOT IN ('male', 'female')")

	return nil
}
