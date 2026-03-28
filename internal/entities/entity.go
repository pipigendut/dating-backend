package entities


type EntityType string

const (
	EntityTypeUser  EntityType = "user"
	EntityTypeGroup EntityType = "group"
)

type Entity struct {
	BaseModel
	Type EntityType `gorm:"type:varchar(20);not null;index" json:"type"`

	// Associations for preloading
	User  *User  `gorm:"foreignKey:EntityID" json:"user,omitempty"`
	Group *Group `gorm:"foreignKey:EntityID" json:"group,omitempty"`
}

func (Entity) TableName() string {
	return "entities"
}
