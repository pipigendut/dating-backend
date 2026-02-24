package seeds

import (
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/entities"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func SeedMasterData(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {

		// 1. Seed Master Genders
		genders := []entities.MasterGender{
			{ID: uuid.MustParse("e0000000-0000-0000-0000-000000000001"), Code: "male", Name: "Men", Icon: "👨", IsActive: true},
			{ID: uuid.MustParse("e0000000-0000-0000-0000-000000000002"), Code: "female", Name: "Women", Icon: "👩", IsActive: true},
			{ID: uuid.MustParse("e0000000-0000-0000-0000-000000000003"), Code: "everyone", Name: "Everyone", Icon: "🌍", IsActive: true},
		}

		for _, g := range genders {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"code", "name", "icon", "is_active"}),
			}).Create(&g).Error; err != nil {
				return err
			}
		}

		// 2. Seed Master Relationship Types
		relations := []entities.MasterRelationshipType{
			{ID: uuid.MustParse("a0000000-0000-0000-0000-000000000001"), Code: "long_term", Name: "Long-term partner", Icon: "💍", IsActive: true},
			{ID: uuid.MustParse("a0000000-0000-0000-0000-000000000002"), Code: "short_term", Name: "Short-term", Icon: "🥂", IsActive: true},
			{ID: uuid.MustParse("a0000000-0000-0000-0000-000000000003"), Code: "short_term_fun", Name: "Short-term fun", Icon: "🎉", IsActive: true},
			{ID: uuid.MustParse("a0000000-0000-0000-0000-000000000004"), Code: "new_friends", Name: "New friends", Icon: "👋", IsActive: true},
			{ID: uuid.MustParse("a0000000-0000-0000-0000-000000000005"), Code: "figuring_out", Name: "Still figuring it out", Icon: "🤔", IsActive: true},
		}

		for _, r := range relations {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"code", "name", "icon", "is_active"}),
			}).Create(&r).Error; err != nil {
				return err
			}
		}

		// 3. Seed Master Languages
		languages := []entities.MasterLanguage{
			{ID: uuid.MustParse("b0000000-0000-0000-0000-000000000001"), Code: "en", Name: "English", Icon: "🇺🇸", IsActive: true},
			{ID: uuid.MustParse("b0000000-0000-0000-0000-000000000002"), Code: "id", Name: "Indonesian", Icon: "🇮🇩", IsActive: true},
			{ID: uuid.MustParse("b0000000-0000-0000-0000-000000000003"), Code: "zh", Name: "Chinese", Icon: "🇨🇳", IsActive: true},
			{ID: uuid.MustParse("b0000000-0000-0000-0000-000000000004"), Code: "ja", Name: "Japanese", Icon: "🇯🇵", IsActive: true},
			{ID: uuid.MustParse("b0000000-0000-0000-0000-000000000005"), Code: "ko", Name: "Korean", Icon: "🇰🇷", IsActive: true},
			{ID: uuid.MustParse("b0000000-0000-0000-0000-000000000006"), Code: "es", Name: "Spanish", Icon: "🇪🇸", IsActive: true},
			{ID: uuid.MustParse("b0000000-0000-0000-0000-000000000007"), Code: "fr", Name: "French", Icon: "🇫🇷", IsActive: true},
			{ID: uuid.MustParse("b0000000-0000-0000-0000-000000000008"), Code: "de", Name: "German", Icon: "🇩🇪", IsActive: true},
		}

		for _, l := range languages {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"code", "name", "icon", "is_active"}),
			}).Create(&l).Error; err != nil {
				return err
			}
		}

		// 4. Seed Master Interests (Interests usually don't have a unique 'code', we can conflict handle on 'id' if preset)
		interests := []entities.MasterInterest{
			{ID: uuid.MustParse("c0000000-0000-0000-0000-000000000001"), Name: "Travel", Icon: "✈️", IsActive: true},
			{ID: uuid.MustParse("c0000000-0000-0000-0000-000000000002"), Name: "Music", Icon: "🎵", IsActive: true},
			{ID: uuid.MustParse("c0000000-0000-0000-0000-000000000003"), Name: "Food", Icon: "🍕", IsActive: true},
			{ID: uuid.MustParse("c0000000-0000-0000-0000-000000000004"), Name: "Sports", Icon: "⚽", IsActive: true},
			{ID: uuid.MustParse("c0000000-0000-0000-0000-000000000005"), Name: "Photography", Icon: "📸", IsActive: true},
			{ID: uuid.MustParse("c0000000-0000-0000-0000-000000000006"), Name: "Art", Icon: "🎨", IsActive: true},
			{ID: uuid.MustParse("c0000000-0000-0000-0000-000000000007"), Name: "Gaming", Icon: "🎮", IsActive: true},
			{ID: uuid.MustParse("c0000000-0000-0000-0000-000000000008"), Name: "Reading", Icon: "📚", IsActive: true},
			{ID: uuid.MustParse("c0000000-0000-0000-0000-000000000009"), Name: "Movies", Icon: "🎬", IsActive: true},
			{ID: uuid.MustParse("c0000000-0000-0000-0000-000000000010"), Name: "Fitness", Icon: "💪", IsActive: true},
		}

		for _, i := range interests {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"name", "icon", "is_active"}),
			}).Create(&i).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
