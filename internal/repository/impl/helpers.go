package impl

import (
	"gorm.io/gorm"
)

// ApplyFullUserPreload applies all standard preloads for a complete User Profile.
// The prefix parameter is used for nested relationships (e.g., "Members.User.").
func ApplyFullUserPreload(db *gorm.DB, prefix string) *gorm.DB {
	if prefix != "" && prefix[len(prefix)-1] != '.' {
		prefix += "."
	}

	return db.
		Preload(prefix + "Photos", func(db *gorm.DB) *gorm.DB { return db.Order("is_main DESC, created_at ASC") }).
		Preload(prefix + "Gender").
		Preload(prefix + "RelationshipType").
		Preload(prefix + "InterestedGenders").
		Preload(prefix + "Interests").
		Preload(prefix + "Languages").
		Preload(prefix + "Subscriptions", "is_active = true").
		Preload(prefix + "Subscriptions.Plan").
		Preload(prefix + "Consumables")
}
