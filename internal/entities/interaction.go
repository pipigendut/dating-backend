package entities

import (
	"time"

	"github.com/google/uuid"
)

type SwipeDirection string

const (
	SwipeDirectionLike  SwipeDirection = "LIKE"
	SwipeDirectionPass  SwipeDirection = "DISLIKE"
	SwipeDirectionCrush SwipeDirection = "CRUSH"
)

type Swipe struct {
	BaseModel
	SwiperEntityID uuid.UUID      `gorm:"type:uuid;not null;index;uniqueIndex:idx_entity_swipe" json:"swiper_entity_id"`
	SwipedEntityID uuid.UUID      `gorm:"type:uuid;not null;index;uniqueIndex:idx_entity_swipe" json:"swiped_entity_id"`
	Direction      SwipeDirection `gorm:"type:varchar(20);not null;index" json:"direction"`
	IsBoosted      bool           `gorm:"default:false;index" json:"is_boosted"`
	RankingScore   float64        `gorm:"default:0;index" json:"ranking_score"`
	PriorityScore  int            `gorm:"default:0;index" json:"priority_score"`
	ProcessedAt    *time.Time     `gorm:"index" json:"processed_at,omitempty"`

	// Associations
	SwiperEntity *Entity `gorm:"foreignKey:SwiperEntityID" json:"swiper_entity,omitempty"`
	SwipedEntity *Entity `gorm:"foreignKey:SwipedEntityID" json:"swiped_entity,omitempty"`
}

type Match struct {
	BaseModel
	Entity1ID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_entity_pair" json:"entity1_id"`
	Entity2ID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_entity_pair" json:"entity2_id"`

	// Associations
	Entity1 *Entity `gorm:"foreignKey:Entity1ID" json:"entity1,omitempty"`
	Entity2 *Entity `gorm:"foreignKey:Entity2ID" json:"entity2,omitempty"`
}

type EntityBoost struct {
	BaseModel
	EntityID  uuid.UUID `gorm:"type:uuid;not null;index" json:"entity_id"`
	StartedAt time.Time `gorm:"index" json:"started_at"`
	ExpiresAt time.Time `gorm:"index" json:"expires_at"`

	// Associations
	Entity *Entity `gorm:"foreignKey:EntityID" json:"entity,omitempty"`
}

type EntityImpression struct {
	BaseModel
	ViewerEntityID uuid.UUID `gorm:"type:uuid;not null;index" json:"viewer_entity_id"`
	ShownEntityID  uuid.UUID `gorm:"type:uuid;not null;index" json:"shown_entity_id"`
	ShownAt        time.Time `gorm:"index;default:now()" json:"shown_at"`
}

type EntityUnmatch struct {
	BaseModel
	SwiperEntityID uuid.UUID `gorm:"type:uuid;not null;index" json:"swiper_entity_id"`
	TargetEntityID uuid.UUID `gorm:"type:uuid;not null;index" json:"target_entity_id"`
}

func (Swipe) TableName() string {
	return "swipes"
}

func (Match) TableName() string {
	return "matches"
}

func (EntityBoost) TableName() string {
	return "entity_boosts"
}

func (EntityImpression) TableName() string {
	return "entity_impressions"
}

func (EntityUnmatch) TableName() string {
	return "entity_unmatches"
}
