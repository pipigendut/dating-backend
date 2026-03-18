package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserStatus string

const (
	UserStatusOnboarding UserStatus = "onboarding"
	UserStatusActive     UserStatus = "active"
	UserStatusBanned     UserStatus = "banned"
)

type User struct {
	ID                 uuid.UUID `gorm:"primaryKey;type:uuid"`
	Email              *string   `gorm:"uniqueIndex"`
	PasswordHash       *string
	FullName           string
	DateOfBirth        time.Time `gorm:"index"`
	HeightCM           int
	Bio                string
	LocationCity       string
	LocationCountry    string
	Latitude           *float64
	Longitude          *float64
	GenderID           *uuid.UUID `gorm:"type:uuid;index"`
	RelationshipTypeID *uuid.UUID `gorm:"type:uuid;index"`
	IsPremium          bool       `gorm:"default:false"`
	LastActiveAt       time.Time  `gorm:"index"`
	Status             UserStatus `gorm:"index"`
	SwipeCountToday    int        `gorm:"default:0"`
	CreatedAt          time.Time  `gorm:"autoCreateTime"`
	UpdatedAt          time.Time  `gorm:"autoUpdateTime"`

	// Associations
	Gender           *MasterGender           `gorm:"foreignKey:GenderID"`
	RelationshipType *MasterRelationshipType `gorm:"foreignKey:RelationshipTypeID"`
	Photos           []Photo                 `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
	AuthProviders    []AuthProvider          `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
	Devices          []Device                `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
	RefreshTokens    []RefreshToken          `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
	Subscriptions    []UserSubscription      `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
	Consumables      []UserConsumable        `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`

	// Many-to-Many Associations via pivot tables
	InterestedGenders []MasterGender   `gorm:"many2many:user_interested_genders;joinForeignKey:user_id;joinReferences:gender_id;constraint:OnDelete:CASCADE"`
	Interests         []MasterInterest `gorm:"many2many:user_interests;joinForeignKey:user_id;joinReferences:interest_id;constraint:OnDelete:CASCADE"`
	Languages         []MasterLanguage `gorm:"many2many:user_languages;joinForeignKey:user_id;joinReferences:language_id;constraint:OnDelete:CASCADE"`
}

func (u *User) GetMainPhotoProfile() *Photo {
	if u == nil || len(u.Photos) == 0 {
		return nil
	}
	for i := range u.Photos {
		if u.Photos[i].IsMain {
			return &u.Photos[i]
		}
	}
	return &u.Photos[0]
}

type AuthProvider struct {
	ID             uuid.UUID `gorm:"primaryKey;type:uuid"`
	UserID         uuid.UUID `gorm:"type:uuid;index"`
	Provider       string    `gorm:"uniqueIndex:idx_provider_user"`
	ProviderUserID string    `gorm:"uniqueIndex:idx_provider_user"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
}

type Photo struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid"`
	UserID    uuid.UUID `gorm:"type:uuid;index"`
	URL       string
	IsMain    bool
	SortOrder int
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type Device struct {
	ID          uuid.UUID `gorm:"primaryKey;type:uuid"`
	UserID      uuid.UUID `gorm:"type:uuid;index;uniqueIndex:idx_device_user"`
	DeviceID    string    `gorm:"uniqueIndex:idx_device_user"`
	DeviceName  string
	DeviceModel string
	OSVersion   string
	AppVersion  string
	FCMToken    *string
	LastIP      string
	LastLogin   time.Time
	IsActive    bool
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

type RefreshToken struct {
	ID        uuid.UUID  `gorm:"primaryKey;type:uuid"`
	UserID    uuid.UUID  `gorm:"type:uuid;index"`
	DeviceID  uuid.UUID  `gorm:"type:uuid;index"`
	TokenHash string     `gorm:"uniqueIndex"`
	ExpiresAt time.Time  `gorm:"index"`
	RevokedAt *time.Time `gorm:"index"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
}

// Master Tables

type MasterGender struct {
	ID       uuid.UUID `gorm:"primaryKey;type:uuid"`
	Code     string    `gorm:"uniqueIndex"`
	Name     string
	Icon     string
	IsActive bool `gorm:"default:true"`
}

type MasterRelationshipType struct {
	ID       uuid.UUID `gorm:"primaryKey;type:uuid"`
	Code     string    `gorm:"uniqueIndex"`
	Name     string
	Icon     string
	IsActive bool `gorm:"default:true"`
}

type MasterInterest struct {
	ID       uuid.UUID `gorm:"primaryKey;type:uuid"`
	Name     string
	Icon     string
	IsActive bool `gorm:"default:true"`
}

type MasterLanguage struct {
	ID       uuid.UUID `gorm:"primaryKey;type:uuid"`
	Code     string    `gorm:"uniqueIndex"`
	Name     string
	Icon     string
	IsActive bool `gorm:"default:true"`
}

// Pivot Tables

type UserInterestedGender struct {
	UserID   uuid.UUID `gorm:"primaryKey;type:uuid;index"`
	GenderID uuid.UUID `gorm:"primaryKey;type:uuid;index"`
}

type UserInterest struct {
	UserID     uuid.UUID `gorm:"primaryKey;type:uuid;index"`
	InterestID uuid.UUID `gorm:"primaryKey;type:uuid;index"`
}

type UserLanguage struct {
	UserID     uuid.UUID `gorm:"primaryKey;type:uuid;index"`
	LanguageID uuid.UUID `gorm:"primaryKey;type:uuid;index"`
}

// Matching System

type SwipeDirection string

const (
	SwipeDirectionLike    SwipeDirection = "LIKE"
	SwipeDirectionDislike SwipeDirection = "DISLIKE"
	SwipeDirectionCrush   SwipeDirection = "CRUSH"
)

type Swipe struct {
	ID           uuid.UUID      `gorm:"primaryKey;type:uuid"`
	SwiperID     uuid.UUID      `gorm:"type:uuid;uniqueIndex:idx_swiper_swiped;index:idx_swiper_direction;index"`
	SwipedID     uuid.UUID      `gorm:"type:uuid;uniqueIndex:idx_swiper_swiped;index:idx_swiped_direction;index"`
	Direction     SwipeDirection `gorm:"type:varchar(20);index;index:idx_swiper_direction;index:idx_swiped_direction"`
	IsBoosted     bool           `gorm:"default:false;index"`
	RankingScore  float64        `gorm:"index"` // Algorithm-based score for ranking
	PriorityScore int            `gorm:"default:0;index"`
	ProcessedAt   *time.Time     `gorm:"index"` // For anti-fast match delay
	CreatedAt     time.Time      `gorm:"autoCreateTime;index"`
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

type Match struct {
	ID        uuid.UUID      `gorm:"primaryKey;type:uuid"`
	// Use deterministic IDs: UserLowID always < UserHighID to prevent duplicates
	UserLowID  uuid.UUID      `gorm:"type:uuid;uniqueIndex:idx_user_pair;index"`
	UserHighID uuid.UUID      `gorm:"type:uuid;uniqueIndex:idx_user_pair;index"`
	VisibleAt  time.Time      `gorm:"index"` // When this match becomes visible to users
	CreatedAt  time.Time      `gorm:"autoCreateTime;index"`
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

type Unmatch struct {
	ID           uuid.UUID `gorm:"primaryKey;type:uuid"`
	UserID       uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_user_target;index;not null"`
	TargetUserID uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_user_target;index;not null"`
	MatchID      uuid.UUID `gorm:"type:uuid;index;not null"`
	CreatedAt    time.Time `gorm:"autoCreateTime;index"`
}

type UserImpression struct {
	ID          uuid.UUID `gorm:"primaryKey;type:uuid"`
	ViewerID    uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_viewer_shown;index"`
	ShownUserID uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_viewer_shown;index"`
	ShownAt     time.Time `gorm:"autoCreateTime;index"`
}

type UserBoost struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid"`
	UserID    uuid.UUID `gorm:"type:uuid;index;not null"`
	StartedAt time.Time `gorm:"index"`
	ExpiredAt time.Time `gorm:"index"`
	IsActive  bool      `gorm:"default:true;index"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type AppConfig struct {
	Key         string    `gorm:"primaryKey"`
	Value       string    `gorm:"type:jsonb"`
	Description string
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

// Monetization Entities

type SubscriptionPlan struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid" json:"id"`
	Name      string    `gorm:"uniqueIndex;not null" json:"name"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Features []SubscriptionPlanFeature `gorm:"foreignKey:PlanID;constraint:OnDelete:CASCADE" json:"features"`
	Prices   []SubscriptionPrice        `gorm:"foreignKey:PlanID;constraint:OnDelete:CASCADE" json:"prices"`
}

type SubscriptionPrice struct {
	ID           uuid.UUID `gorm:"primaryKey;type:uuid" json:"id"`
	PlanID       uuid.UUID `gorm:"type:uuid;index;not null" json:"plan_id"`
	DurationType string    `gorm:"index;not null" json:"duration_type"` // weekly, monthly, quarterly, yearly
	Price        float64   `gorm:"not null" json:"price"`
	Currency     string    `gorm:"type:varchar(10);default:'USD'" json:"currency"`
	ExternalSlug string    `gorm:"uniqueIndex" json:"external_slug"` // For App Store / Play Store Product ID
}

type SubscriptionPlanFeature struct {
	ID           uuid.UUID `gorm:"primaryKey;type:uuid" json:"id"`
	PlanID       uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_plan_feature;not null" json:"plan_id"`
	FeatureKey   string    `gorm:"uniqueIndex:idx_plan_feature;not null" json:"feature_key"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	Category     string    `gorm:"type:varchar(50);default:'General'" json:"category"`
	Icon         string    `json:"icon"`
	DisplayTitle string    `json:"display_title"`
	IsConsumable bool      `gorm:"default:false" json:"is_consumable"`
	Amount       int       `gorm:"default:0" json:"amount"`
}

type UserSubscription struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	PlanID    uuid.UUID `gorm:"type:uuid;index;not null" json:"plan_id"`
	StartedAt time.Time `gorm:"index" json:"started_at"`
	ExpiredAt time.Time `gorm:"index" json:"expired_at"`
	IsActive  bool      `gorm:"default:true;index" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`

	Plan *SubscriptionPlan `gorm:"foreignKey:PlanID" json:"plan,omitempty"`
}

type UserConsumable struct {
	ID        uuid.UUID  `gorm:"primaryKey;type:uuid" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`
	Type      string     `gorm:"index;not null" json:"type"` // boost, crush
	Remaining int        `gorm:"default:0" json:"remaining"`
	ExpiredAt *time.Time `gorm:"index" json:"expired_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type ConsumableItem struct {
	ID           uuid.UUID `gorm:"primaryKey;type:uuid" json:"id"`
	ItemType     string    `gorm:"index;not null" json:"item_type"` // boost, crush
	Amount       int       `gorm:"not null" json:"amount"`
	Price        float64   `gorm:"not null" json:"price"`
	Currency     string    `gorm:"type:varchar(10);default:'USD'" json:"currency"`
	ExternalSlug string    `gorm:"uniqueIndex" json:"external_slug"`
}

func (Swipe) TableName() string {
	return "swipes"
}

func (Match) TableName() string {
	return "matches"
}
