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
	SwipeDirectionCrush   SwipeDirection = "CRUSH"
)

type Swipe struct {
	ID           uuid.UUID      `gorm:"primaryKey;type:uuid"`
	SwiperID     uuid.UUID      `gorm:"type:uuid;uniqueIndex:idx_swiper_swiped;index:idx_swiper_direction;index"`
	SwipedID     uuid.UUID      `gorm:"type:uuid;uniqueIndex:idx_swiper_swiped;index:idx_swiped_direction;index"`
	Direction    SwipeDirection `gorm:"type:varchar(20);index;index:idx_swiper_direction;index:idx_swiped_direction"`
	IsBoosted    bool           `gorm:"default:false;index"`
	RankingScore float64        `gorm:"index"` // Algorithm-based score for ranking
	CreatedAt    time.Time      `gorm:"autoCreateTime;index"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

type Match struct {
	ID        uuid.UUID      `gorm:"primaryKey;type:uuid"`
	// Use deterministic IDs: UserLowID always < UserHighID to prevent duplicates
	UserLowID  uuid.UUID      `gorm:"type:uuid;uniqueIndex:idx_user_pair;index"`
	UserHighID uuid.UUID      `gorm:"type:uuid;uniqueIndex:idx_user_pair;index"`
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
	Key   string `gorm:"primaryKey"`
	Value string
}
