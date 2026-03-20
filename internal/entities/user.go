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
	BaseModel
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
	Age                int        `gorm:"index"`
	SwipeCountToday    int        `gorm:"default:0"`
	VerifiedAt         *time.Time `gorm:"index"`

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

func (u *User) BeforeSave(tx *gorm.DB) (err error) {
	if !u.DateOfBirth.IsZero() {
		now := time.Now()
		age := now.Year() - u.DateOfBirth.Year()
		if now.Month() < u.DateOfBirth.Month() || (now.Month() == u.DateOfBirth.Month() && now.Day() < u.DateOfBirth.Day()) {
			age--
		}
		u.Age = age
	}
	return
}

type AuthProvider struct {
	BaseModel
	UserID         uuid.UUID `gorm:"type:uuid;index"`
	Provider       string    `gorm:"uniqueIndex:idx_provider_user"`
	ProviderUserID string    `gorm:"uniqueIndex:idx_provider_user"`
}

type Photo struct {
	BaseModel
	UserID    uuid.UUID `gorm:"type:uuid;index"`
	URL       string
	IsMain    bool
	SortOrder int
}

type Device struct {
	BaseModel
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
}

type RefreshToken struct {
	BaseModel
	UserID    uuid.UUID  `gorm:"type:uuid;index"`
	DeviceID  uuid.UUID  `gorm:"type:uuid;index"`
	TokenHash string     `gorm:"uniqueIndex"`
	ExpiresAt time.Time  `gorm:"index"`
	RevokedAt *time.Time `gorm:"index"`
}

// Master Tables

type MasterGender struct {
	BaseModel
	Code     string    `gorm:"uniqueIndex"`
	Name     string
	Icon     string
	IsActive bool `gorm:"default:true"`
}

type MasterRelationshipType struct {
	BaseModel
	Code     string    `gorm:"uniqueIndex"`
	Name     string
	Icon     string
	IsActive bool `gorm:"default:true"`
}

type MasterInterest struct {
	BaseModel
	Name     string
	Icon     string
	IsActive bool `gorm:"default:true"`
}

type MasterLanguage struct {
	BaseModel
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
	SoftDeleteModel
	SwiperID     uuid.UUID      `gorm:"type:uuid;uniqueIndex:idx_swiper_swiped;index:idx_swiper_direction;index"`
	SwipedID     uuid.UUID      `gorm:"type:uuid;uniqueIndex:idx_swiper_swiped;index:idx_swiped_direction;index"`
	Direction     SwipeDirection `gorm:"type:varchar(20);index;index:idx_swiper_direction;index:idx_swiped_direction"`
	IsBoosted     bool           `gorm:"default:false;index"`
	RankingScore  float64        `gorm:"index"` // Algorithm-based score for ranking
	PriorityScore int            `gorm:"default:0;index"`
	ProcessedAt   *time.Time     `gorm:"index"` // For anti-fast match delay
}

type Match struct {
	SoftDeleteModel
	// Use deterministic IDs: UserLowID always < UserHighID to prevent duplicates
	UserLowID  uuid.UUID      `gorm:"type:uuid;uniqueIndex:idx_user_pair;index"`
	UserHighID uuid.UUID      `gorm:"type:uuid;uniqueIndex:idx_user_pair;index"`
	VisibleAt  time.Time      `gorm:"index"` // When this match becomes visible to users
}

type Unmatch struct {
	SoftDeleteModel
	UserID       uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_user_target;index;not null"`
	TargetUserID uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_user_target;index;not null"`
	MatchID      uuid.UUID `gorm:"type:uuid;index;not null"`
}

type UserImpression struct {
	BaseModel
	ViewerID    uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_viewer_shown;index"`
	ShownUserID uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_viewer_shown;index"`
}

type UserBoost struct {
	BaseModel
	UserID    uuid.UUID `gorm:"type:uuid;index;not null"`
	StartedAt time.Time `gorm:"index"`
	ExpiredAt time.Time `gorm:"index"`
	IsActive  bool      `gorm:"default:true;index"`
}

type AppConfig struct {
	BaseModel
	Key         string    `gorm:"uniqueIndex"`
	Value       string    `gorm:"type:jsonb"`
	Description string
}

// Monetization Entities

type SubscriptionPlan struct {
	BaseModel
	Name      string    `gorm:"uniqueIndex;not null" json:"name"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`

	Features []SubscriptionPlanFeature `gorm:"foreignKey:PlanID;constraint:OnDelete:CASCADE" json:"features"`
	Prices   []SubscriptionPrice        `gorm:"foreignKey:PlanID;constraint:OnDelete:CASCADE" json:"prices"`
}

type SubscriptionPrice struct {
	BaseModel
	PlanID       uuid.UUID `gorm:"type:uuid;index;not null" json:"plan_id"`
	DurationType string    `gorm:"index;not null" json:"duration_type"` // weekly, monthly, quarterly, yearly
	Price        float64   `gorm:"not null" json:"price"`
	Currency     string    `gorm:"type:varchar(10);default:'USD'" json:"currency"`
	ExternalSlug string    `gorm:"uniqueIndex" json:"external_slug"` // For App Store / Play Store Product ID
}

type SubscriptionPlanFeature struct {
	BaseModel
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
	SoftDeleteModel
	UserID    uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_user_subs;not null" json:"user_id"`
	PlanID    uuid.UUID `gorm:"type:uuid;index;not null" json:"plan_id"`
	StartedAt time.Time `gorm:"index" json:"started_at"`
	ExpiredAt time.Time `gorm:"index" json:"expired_at"`
	IsActive  bool      `gorm:"default:true;index" json:"is_active"`

	Plan *SubscriptionPlan `gorm:"foreignKey:PlanID" json:"plan,omitempty"`
}

type UserConsumable struct {
	BaseModel
	UserID     uuid.UUID  `gorm:"type:uuid;uniqueIndex:idx_user_cons;not null" json:"user_id"`
	ItemType   string     `gorm:"index;uniqueIndex:idx_user_cons;not null" json:"item_type"` // boost, crush
	Amount     int        `gorm:"default:0" json:"amount"`
	LastUsedAt *time.Time `gorm:"index" json:"last_used_at,omitempty"`
}

type ConsumablePackage struct {
	BaseModel
	Name     string  `json:"name"`
	ItemType string  `gorm:"index;not null" json:"item_type"` // boost, crush
	Amount   int     `gorm:"not null" json:"amount"`
	Price    float64 `gorm:"not null" json:"price"`
	IsActive bool    `gorm:"default:true" json:"is_active"`
}

func (Swipe) TableName() string {
	return "swipes"
}

func (Match) TableName() string {
	return "matches"
}
