package v1

import (
dtov1 "github.com/pipigendut/dating-backend/internal/delivery/http/dto/v1"
	"net/http"

	"github.com/gin-gonic/gin"
	base "github.com/pipigendut/dating-backend/internal/delivery/http/dto"
	"github.com/pipigendut/dating-backend/internal/services"
)

type MasterHandler struct {
	masterService *services.MasterService
	adService     services.AdvertisementService
}

func NewMasterHandler(masterService *services.MasterService, adService services.AdvertisementService) *MasterHandler {
	return &MasterHandler{masterService: masterService, adService: adService}
}

// GetGenders godoc
// @Summary      Get all active genders
// @Description  Returns a list of all active gender options for user profiling.
// @Tags         master
// @Produce      json
// @Success      200      {object}  base.BaseResponse{data=[]dtov1.MasterItemResponse}
// @Failure      500      {object}  base.BaseResponse
// @Router       /master/genders [get]
func (h *MasterHandler) GetGenders(c *gin.Context) {
	items, err := h.masterService.GetGenders()
	if err != nil {
		base.Error(c, http.StatusInternalServerError, "Failed to retrieve genders", err.Error())
		return
	}
	base.OK(c, ToMasterItemsResponseGenders(items))
}

// GetRelationshipTypes godoc
// @Summary      Get all active relationship types
// @Description  Returns a list of all active relationship types (e.g., Short-term, Long-term) for user profiling.
// @Tags         master
// @Produce      json
// @Success      200      {object}  base.BaseResponse{data=[]dtov1.MasterItemResponse}
// @Failure      500      {object}  base.BaseResponse
// @Router       /master/relationship-types [get]
func (h *MasterHandler) GetRelationshipTypes(c *gin.Context) {
	items, err := h.masterService.GetRelationshipTypes()
	if err != nil {
		base.Error(c, http.StatusInternalServerError, "Failed to retrieve relationship types", err.Error())
		return
	}
	base.OK(c, ToMasterItemsResponseRelationshipTypes(items))
}

// GetInterests godoc
// @Summary      Get all active interests
// @Description  Returns a list of all active interests/hobbies options.
// @Tags         master
// @Produce      json
// @Success      200      {object}  base.BaseResponse{data=[]dtov1.MasterItemResponse}
// @Failure      500      {object}  base.BaseResponse
// @Router       /master/interests [get]
func (h *MasterHandler) GetInterests(c *gin.Context) {
	items, err := h.masterService.GetInterests()
	if err != nil {
		base.Error(c, http.StatusInternalServerError, "Failed to retrieve interests", err.Error())
		return
	}
	base.OK(c, ToMasterItemsResponseInterests(items))
}

// GetLanguages godoc
// @Summary      Get all active languages
// @Description  Returns a list of all active spoken languages options.
// @Tags         master
// @Produce      json
// @Success      200      {object}  base.BaseResponse{data=[]dtov1.MasterItemResponse}
// @Failure      500      {object}  base.BaseResponse
// @Router       /master/languages [get]
func (h *MasterHandler) GetLanguages(c *gin.Context) {
	items, err := h.masterService.GetLanguages()
	if err != nil {
		base.Error(c, http.StatusInternalServerError, "Failed to retrieve languages", err.Error())
		return
	}
	base.OK(c, ToMasterItemsResponseLanguages(items))
}

func ToMasterItemsResponseGenders(genders []services.GenderDTO) []dtov1.MasterItemResponse {
	var results []dtov1.MasterItemResponse
	for _, g := range genders {
		results = append(results, dtov1.MasterItemResponse{ID: g.ID, Name: g.Name, Icon: g.Icon})
	}
	return results
}

func ToMasterItemsResponseRelationshipTypes(types []services.RelationshipTypeDTO) []dtov1.MasterItemResponse {
	var results []dtov1.MasterItemResponse
	for _, t := range types {
		results = append(results, dtov1.MasterItemResponse{ID: t.ID, Name: t.Name, Icon: t.Icon})
	}
	return results
}

func ToMasterItemsResponseInterests(interests []services.InterestDTO) []dtov1.MasterItemResponse {
	var results []dtov1.MasterItemResponse
	for _, i := range interests {
		results = append(results, dtov1.MasterItemResponse{ID: i.ID, Name: i.Name, Icon: i.Icon})
	}
	return results
}

func ToMasterItemsResponseLanguages(languages []services.LanguageDTO) []dtov1.MasterItemResponse {
	var results []dtov1.MasterItemResponse
	for _, l := range languages {
		results = append(results, dtov1.MasterItemResponse{ID: l.ID, Name: l.Name, Icon: l.Icon})
	}
	return results
}
