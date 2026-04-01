package entities

// Advertisement represents a promotional ad in the application
type Advertisement struct {
	BaseModel
	Source    string `gorm:"type:varchar(50);not null;index" json:"source"`    // internal, sponsor, admob
	Placement string `gorm:"type:varchar(50);not null;index" json:"placement"` // carousel, card_deck, popup_modal, interstitial
	ImageURL  string `gorm:"type:text" json:"image_url"`
	Link      string `gorm:"type:text" json:"link"`
	Sponsor   string `gorm:"type:varchar(255)" json:"sponsor"`
	IsActive  bool   `gorm:"default:true;index" json:"is_active"`
	SortOrder int    `gorm:"default:0;index" json:"sort_order"`
}
