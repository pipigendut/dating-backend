package entities

import (
	"time"

	"github.com/google/uuid"
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
	Status             UserStatus `gorm:"index"`
	CreatedAt          time.Time  `gorm:"autoCreateTime"`
	UpdatedAt          time.Time  `gorm:"autoUpdateTime"`

	// Associations
	Gender           *MasterGender           `gorm:"foreignKey:GenderID"`
	RelationshipType *MasterRelationshipType `gorm:"foreignKey:RelationshipTypeID"`
	Photos           []Photo                 `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
	AuthProviders    []AuthProvider          `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
	Devices          []Device                `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
	RefreshTokens    []RefreshToken          `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`

	// Many-to-Many Associations via pivot tables
	InterestedGenders []MasterGender   `gorm:"many2many:user_interested_genders;joinForeignKey:user_id;joinReferences:gender_id;constraint:OnDelete:CASCADE"`
	Interests         []MasterInterest `gorm:"many2many:user_interests;joinForeignKey:user_id;joinReferences:interest_id;constraint:OnDelete:CASCADE"`
	Languages         []MasterLanguage `gorm:"many2many:user_languages;joinForeignKey:user_id;joinReferences:language_id;constraint:OnDelete:CASCADE"`
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
	UserID      uuid.UUID `gorm:"type:uuid;index"`
	DeviceID    string    `gorm:"uniqueIndex"`
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
)

type Swipe struct {
	ID        uuid.UUID      `gorm:"primaryKey;type:uuid"`
	SwiperID  uuid.UUID      `gorm:"type:uuid;index"`
	SwipedID  uuid.UUID      `gorm:"type:uuid;index"`
	Direction SwipeDirection `gorm:"index"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
}

type Match struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid"`
	User1ID   uuid.UUID `gorm:"type:uuid;index"`
	User2ID   uuid.UUID `gorm:"type:uuid;index"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
