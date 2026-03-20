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
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("e0000000-0000-0000-0000-000000000001")}, Code: "male", Name: "Men", Icon: "👨", IsActive: true},
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("e0000000-0000-0000-0000-000000000002")}, Code: "female", Name: "Women", Icon: "👩", IsActive: true},
		}

		for _, g := range genders {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"code", "name", "icon", "is_active", "updated_at"}),
			}).Create(&g).Error; err != nil {
				return err
			}
		}

		// 2. Seed Master Relationship Types
		relations := []entities.MasterRelationshipType{
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("a0000000-0000-0000-0000-000000000001")}, Code: "long_term", Name: "Long-term partner", Icon: "💍", IsActive: true},
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("a0000000-0000-0000-0000-000000000002")}, Code: "short_term", Name: "Short-term", Icon: "🥂", IsActive: true},
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("a0000000-0000-0000-0000-000000000003")}, Code: "short_term_fun", Name: "Short-term fun", Icon: "🎉", IsActive: true},
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("a0000000-0000-0000-0000-000000000004")}, Code: "new_friends", Name: "New friends", Icon: "👋", IsActive: true},
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("a0000000-0000-0000-0000-000000000005")}, Code: "figuring_out", Name: "Still figuring it out", Icon: "🤔", IsActive: true},
		}

		for _, r := range relations {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"code", "name", "icon", "is_active", "updated_at"}),
			}).Create(&r).Error; err != nil {
				return err
			}
		}

		// 3. Seed Master Languages
		languages := []entities.MasterLanguage{
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("b0000000-0000-0000-0000-000000000001")}, Code: "en", Name: "English", Icon: "🇺🇸", IsActive: true},
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("b0000000-0000-0000-0000-000000000002")}, Code: "id", Name: "Indonesian", Icon: "🇮🇩", IsActive: true},
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("b0000000-0000-0000-0000-000000000003")}, Code: "zh", Name: "Chinese", Icon: "🇨🇳", IsActive: true},
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("b0000000-0000-0000-0000-000000000004")}, Code: "ja", Name: "Japanese", Icon: "🇯🇵", IsActive: true},
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("b0000000-0000-0000-0000-000000000005")}, Code: "ko", Name: "Korean", Icon: "🇰🇷", IsActive: true},
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("b0000000-0000-0000-0000-000000000006")}, Code: "es", Name: "Spanish", Icon: "🇪🇸", IsActive: true},
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("b0000000-0000-0000-0000-000000000007")}, Code: "fr", Name: "French", Icon: "🇫🇷", IsActive: true},
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("b0000000-0000-0000-0000-000000000008")}, Code: "de", Name: "German", Icon: "🇩🇪", IsActive: true},
		}

		for _, l := range languages {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"code", "name", "icon", "is_active", "updated_at"}),
			}).Create(&l).Error; err != nil {
				return err
			}
		}

		// 4. Seed Master Interests (Interests usually don't have a unique 'code', we can conflict handle on 'id' if preset)
		interests := []entities.MasterInterest{
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("c0000000-0000-0000-0000-000000000001")}, Name: "Travel", Icon: "✈️", IsActive: true},
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("c0000000-0000-0000-0000-000000000002")}, Name: "Music", Icon: "🎵", IsActive: true},
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("c0000000-0000-0000-0000-000000000003")}, Name: "Food", Icon: "🍕", IsActive: true},
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("c0000000-0000-0000-0000-000000000004")}, Name: "Sports", Icon: "⚽", IsActive: true},
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("c0000000-0000-0000-0000-000000000005")}, Name: "Photography", Icon: "📸", IsActive: true},
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("c0000000-0000-0000-0000-000000000006")}, Name: "Art", Icon: "🎨", IsActive: true},
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("c0000000-0000-0000-0000-000000000007")}, Name: "Gaming", Icon: "🎮", IsActive: true},
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("c0000000-0000-0000-0000-000000000008")}, Name: "Reading", Icon: "📚", IsActive: true},
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("c0000000-0000-0000-0000-000000000009")}, Name: "Movies", Icon: "🎬", IsActive: true},
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("c0000000-0000-0000-0000-000000000010")}, Name: "Fitness", Icon: "💪", IsActive: true},
		}

		for _, i := range interests {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"name", "icon", "is_active"}),
			}).Create(&i).Error; err != nil {
				return err
			}
		}

		// 5. Seed App Configurations
		configs := []entities.AppConfig{
			{Key: "premium_score", Value: "50"},
			{Key: "boost_score", Value: "200"},
			{Key: "crush_score_bonus", Value: "500"},
			{Key: "swipe_impression_cooldown_premium", Value: "10"},
			{Key: "swipe_impression_cooldown_free", Value: "1"},
			{Key: "swipe_impression_cooldown_boost", Value: `"3"`},
			{Key: "score_weight", Value: `"0.7"`},
			{Key: "random_weight", Value: `"0.3"`},
			{Key: "incoming_like_delay_free", Value: `"60"`},
			{Key: "incoming_like_delay_premium", Value: `"10"`},
			{Key: "dislike_recycle_minutes", Value: `"4320"`}, // 3 days default
			{Key: "max_limit_likes_free", Value: `"50"`},
			{Key: "like_expiry_hours", Value: `"72"`}, // 3 days default
			{Key: "whitelist_emails", Value: `["akbar.maulana090895@gmail.com"]`},
		}

		for _, c := range configs {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "key"}},
				DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
			}).Create(&c).Error; err != nil {
				return err
			}
		}

		// 6. Seed Subscription Plans
		plusID := uuid.MustParse("d0000000-0000-0000-0000-000000000001")
		premiumID := uuid.MustParse("d0000000-0000-0000-0000-000000000002")
		ultimateID := uuid.MustParse("d0000000-0000-0000-0000-000000000003")

		plans := []entities.SubscriptionPlan{
			{BaseModel: entities.BaseModel{ID: plusID}, Name: "Plus", IsActive: true},
			{BaseModel: entities.BaseModel{ID: premiumID}, Name: "Premium", IsActive: true},
			{BaseModel: entities.BaseModel{ID: ultimateID}, Name: "Ultimate", IsActive: true},
		}

		for _, p := range plans {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"name", "is_active", "updated_at"}),
			}).Create(&p).Error; err != nil {
				return err
			}
		}

		// 7. Seed Subscription Plan Features
		featureList := []struct {
			id           string
			planID       uuid.UUID
			featureKey   string
			isActive     bool
			category     string
			icon         string
			displayTitle string
			isConsumable bool
			amount       int
		}{
			{"f0000000-0000-0000-0000-000000000001", plusID, "hide_ads", true, "Take Control", "ShieldOff", "Hide Ads", false, 0},
			{"f0000000-0000-0000-0000-000000000003", plusID, "unlimited_likes", true, "Match+", "Heart", "Unlimited Likes", false, 0},
			{"f0000000-0000-0000-0000-100000000001", plusID, "undo_swipe", true, "Take Control", "RotateCcw", "Undo swipe", false, 0},

			{"f0000000-0000-0000-0000-000000000004", premiumID, "hide_ads", true, "Take Control", "ShieldOff", "Hide Ads", false, 0},
			{"f0000000-0000-0000-0000-000000000005", premiumID, "see_likes", true, "Match+", "Eye", "See Who Likes You", false, 0},
			{"f0000000-0000-0000-0000-000000000006", premiumID, "priority_likes", true, "Match+", "Star", "Priority Likes", false, 0},
			{"f0000000-0000-0000-0000-000000000007", premiumID, "unlimited_likes", true, "Match+", "Heart", "Unlimited Likes", false, 0},
			{"f0000000-0000-0000-0000-100000000002", premiumID, "undo_swipe", true, "Take Control", "RotateCcw", "Undo swipe", false, 0},

			{"f0000000-0000-0000-0000-000000000013", premiumID, "monthly_boost", true, "Take Control", "Zap", "1 free boost per month", true, 1},

			{"f0000000-0000-0000-0000-000000000008", ultimateID, "hide_ads", true, "Take Control", "ShieldOff", "Hide Ads", false, 0},
			{"f0000000-0000-0000-0000-000000000009", ultimateID, "see_likes", true, "Match+", "Eye", "See Who Likes You", false, 0},
			{"f0000000-0000-0000-0000-000000000010", ultimateID, "priority_likes", true, "Match+", "Star", "Priority Likes", false, 0},
			{"f0000000-0000-0000-0000-000000000011", ultimateID, "unlimited_likes", true, "Match+", "Heart", "Unlimited Likes", false, 0},
			{"f0000000-0000-0000-0000-000000000012", ultimateID, "passport_mode", true, "Take Control", "Globe", "Passport Mode", false, 0},
			{"f0000000-0000-0000-0000-000000000014", ultimateID, "monthly_boost", true, "Take Control", "Zap", "1 free boost per month", true, 1},
			{"f0000000-0000-0000-0000-000000000015", ultimateID, "monthly_crush", true, "Match+", "Crown", "5 free crushes per week", true, 5},
			{"f0000000-0000-0000-0000-100000000003", ultimateID, "undo_swipe", true, "Take Control", "RotateCcw", "Undo swipe", false, 0},
		}

		for _, f := range featureList {
			feat := entities.SubscriptionPlanFeature{
				BaseModel:    entities.BaseModel{ID: uuid.MustParse(f.id)},
				PlanID:       f.planID,
				FeatureKey:   f.featureKey,
				IsActive:     f.isActive,
				Category:     f.category,
				Icon:         f.icon,
				DisplayTitle: f.displayTitle,
				IsConsumable: f.isConsumable,
				Amount:       f.amount,
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"is_active", "category", "icon", "display_title", "is_consumable", "amount", "updated_at"}),
			}).Create(&feat).Error; err != nil {
				return err
			}
		}

		// 8. Seed Subscription Prices
		priceList := []entities.SubscriptionPrice{
			// Plus
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("e1000000-0000-0000-0000-000000000001")}, PlanID: plusID, DurationType: "monthly", Price: 9.99, Currency: "USD", ExternalSlug: "plus_1m"},
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("e1000000-0000-0000-0000-000000000002")}, PlanID: plusID, DurationType: "yearly", Price: 59.99, Currency: "USD", ExternalSlug: "plus_1y"},

			// Premium
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("e1000000-0000-0000-0000-000000000003")}, PlanID: premiumID, DurationType: "monthly", Price: 19.99, Currency: "USD", ExternalSlug: "premium_1m"},
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("e1000000-0000-0000-0000-000000000004")}, PlanID: premiumID, DurationType: "yearly", Price: 119.99, Currency: "USD", ExternalSlug: "premium_1y"},

			// Ultimate
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("e1000000-0000-0000-0000-000000000005")}, PlanID: ultimateID, DurationType: "monthly", Price: 29.99, Currency: "USD", ExternalSlug: "ultimate_1m"},
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("e1000000-0000-0000-0000-000000000006")}, PlanID: ultimateID, DurationType: "yearly", Price: 179.99, Currency: "USD", ExternalSlug: "ultimate_1y"},
		}

		for _, pr := range priceList {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"price", "currency", "external_slug", "updated_at"}),
			}).Create(&pr).Error; err != nil {
				return err
			}
		}

		// 9. Seed Consumable Packets
		consumables := []entities.ConsumablePackage{
			// Boosts
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("c1000000-0000-0000-0000-000000000001")}, Name: "1 Boost", ItemType: "boost", Amount: 1, Price: 89000, IsActive: true},
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("c1000000-0000-0000-0000-000000000002")}, Name: "5 Boosts", ItemType: "boost", Amount: 5, Price: 249000, IsActive: true},
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("c1000000-0000-0000-0000-000000000003")}, Name: "15 Boosts", ItemType: "boost", Amount: 15, Price: 499000, IsActive: true},

			// Crushes
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("c1000000-0000-0000-0000-000000000004")}, Name: "3 Crushes", ItemType: "crush", Amount: 3, Price: 49000, IsActive: true},
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("c1000000-0000-0000-0000-000000000005")}, Name: "15 Crushes", ItemType: "crush", Amount: 15, Price: 149000, IsActive: true},
			{BaseModel: entities.BaseModel{ID: uuid.MustParse("c1000000-0000-0000-0000-000000000006")}, Name: "30 Crushes", ItemType: "crush", Amount: 30, Price: 249000, IsActive: true},
		}

		for _, ci := range consumables {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"name", "item_type", "amount", "price", "is_active", "updated_at"}),
			}).Create(&ci).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
