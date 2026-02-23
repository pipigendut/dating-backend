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
	ID            uuid.UUID `gorm:"primaryKey;type:uuid"`
	Email         *string   `gorm:"uniqueIndex"`
	PasswordHash  *string
	Status        UserStatus
	CreatedAt     time.Time      `gorm:"autoCreateTime"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime"`
	Profile       *Profile       `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
	Photos        []Photo        `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
	AuthProviders []AuthProvider `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
	Devices       []Device       `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
	RefreshTokens []RefreshToken `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
}

type AuthProvider struct {
	ID             uuid.UUID `gorm:"primaryKey;type:uuid"`
	UserID         uuid.UUID `gorm:"type:uuid;index"`
	Provider       string    `gorm:"uniqueIndex:idx_provider_user"`
	ProviderUserID string    `gorm:"uniqueIndex:idx_provider_user"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
}

type Profile struct {
	UserID          uuid.UUID `gorm:"primaryKey;type:uuid"`
	FullName        string
	DateOfBirth     time.Time `gorm:"index"`
	Gender          string    `gorm:"index"`
	HeightCM        int
	Bio             string
	InterestedIn    string
	LookingFor      string
	LocationCity    string
	LocationCountry string
	Latitude        *float64
	Longitude       *float64
	Interests       string
	Languages       string
	CreatedAt       time.Time `gorm:"autoCreateTime"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"`
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
