package dtov1

// AdResponse represents the modular ad data structure
type AdResponse struct {
	ID        string `json:"id" example:"uuid"`
	Source    string `json:"source" example:"internal"`    // internal, sponsor, admob
	Placement string `json:"placement" example:"carousel"` // carousel, card_deck, popup_modal, interstitial
	ImageURL  string `json:"image_url,omitempty" example:"https://example.com/banner.jpg"`
	Link      string `json:"link,omitempty" example:"https://example.com/promo"`
	Sponsor   string `json:"sponsor,omitempty" example:"Brand Name"`
	Active    bool   `json:"active" example:"true"`
	Order     int    `json:"order" example:"1"`
}
