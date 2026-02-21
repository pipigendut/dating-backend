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
	ID           uuid.UUID
	Email        *string
	PasswordHash *string
	Status       UserStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Profile      *Profile
	Photos       []Photo
	AuthProviders []AuthProvider
}

type AuthProvider struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	Provider       string
	ProviderUserID string
	CreatedAt      time.Time
}

type Profile struct {
	UserID      uuid.UUID
	FullName    string
	DateOfBirth time.Time
	Gender      string
	HeightCM    int
	Bio         string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Photo struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	URL       string
	IsMain    bool
	SortOrder int
	CreatedAt time.Time
}
