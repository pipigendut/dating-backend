package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	base "github.com/pipigendut/dating-backend/internal/delivery/http/dto"
)

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

// GetAdvertisements godoc
// @Summary      Get active advertisements
// @Description  Returns a list of active ads filtered by placement.
// @Tags         advertisements
// @Accept       json
// @Produce      json
// @Param        placement  query     string  false  "Placement filter (carousel, popup_modal, etc)"
// @Success      200      {object}  base.BaseResponse{data=[]AdResponse}
// @Failure      500      {object}  base.BaseResponse
// @Router       /advertisements [get]
func (h *MasterHandler) GetAdvertisements(c *gin.Context) {
	placement := c.Query("placement")
	ads, err := h.adService.GetActiveAds(placement)
	if err != nil {
		base.Error(c, http.StatusInternalServerError, "Failed to retrieve ads", err.Error())
		return
	}

	var results []AdResponse
	for _, a := range ads {
		results = append(results, AdResponse{
			ID:        a.ID.String(),
			Source:    a.Source,
			Placement: a.Placement,
			ImageURL:  a.ImageURL,
			Link:      a.Link,
			Sponsor:   a.Sponsor,
			Active:    a.IsActive,
			Order:     a.SortOrder,
		})
	}

	base.OK(c, results)
}
