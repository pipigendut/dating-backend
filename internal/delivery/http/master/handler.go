package master

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pipigendut/dating-backend/internal/delivery/http/response"
	"github.com/pipigendut/dating-backend/internal/services"
)

type MasterHandler struct {
	masterService *services.MasterService
	adService     services.AdvertisementService
}

func NewMasterHandler(r *gin.RouterGroup, masterService *services.MasterService, adService services.AdvertisementService) {
	handler := &MasterHandler{masterService: masterService, adService: adService}
	group := r.Group("/master")
	{
		group.GET("/genders", handler.GetGenders)
		group.GET("/relationship-types", handler.GetRelationshipTypes)
		group.GET("/interests", handler.GetInterests)
		group.GET("/languages", handler.GetLanguages)
	}
	r.GET("/advertisements", handler.GetAdvertisements)
}

// GetGenders godoc
// @Summary      Get all active genders
// @Description  Returns a list of all active gender options for user profiling.
// @Tags         master
// @Produce      json
// @Success      200      {object}  response.BaseResponse{data=[]response.MasterItemResponse}
// @Failure      500      {object}  response.BaseResponse
// @Router       /master/genders [get]
func (h *MasterHandler) GetGenders(c *gin.Context) {
	items, err := h.masterService.GetGenders()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve genders", err.Error())
		return
	}
	response.OK(c, ToMasterItemsResponseGenders(items))
}

// GetRelationshipTypes godoc
// @Summary      Get all active relationship types
// @Description  Returns a list of all active relationship types (e.g., Short-term, Long-term) for user profiling.
// @Tags         master
// @Produce      json
// @Success      200      {object}  response.BaseResponse{data=[]response.MasterItemResponse}
// @Failure      500      {object}  response.BaseResponse
// @Router       /master/relationship-types [get]
func (h *MasterHandler) GetRelationshipTypes(c *gin.Context) {
	items, err := h.masterService.GetRelationshipTypes()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve relationship types", err.Error())
		return
	}
	response.OK(c, ToMasterItemsResponseRelationshipTypes(items))
}

// GetInterests godoc
// @Summary      Get all active interests
// @Description  Returns a list of all active interests/hobbies options.
// @Tags         master
// @Produce      json
// @Success      200      {object}  response.BaseResponse{data=[]response.MasterItemResponse}
// @Failure      500      {object}  response.BaseResponse
// @Router       /master/interests [get]
func (h *MasterHandler) GetInterests(c *gin.Context) {
	items, err := h.masterService.GetInterests()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve interests", err.Error())
		return
	}
	response.OK(c, ToMasterItemsResponseInterests(items))
}

// GetLanguages godoc
// @Summary      Get all active languages
// @Description  Returns a list of all active spoken languages options.
// @Tags         master
// @Produce      json
// @Success      200      {object}  response.BaseResponse{data=[]response.MasterItemResponse}
// @Failure      500      {object}  response.BaseResponse
// @Router       /master/languages [get]
func (h *MasterHandler) GetLanguages(c *gin.Context) {
	items, err := h.masterService.GetLanguages()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve languages", err.Error())
		return
	}
	response.OK(c, ToMasterItemsResponseLanguages(items))
}

func ToMasterItemsResponseGenders(genders []services.GenderDTO) []response.MasterItemResponse {
	var results []response.MasterItemResponse
	for _, g := range genders {
		results = append(results, response.MasterItemResponse{ID: g.ID, Name: g.Name, Icon: g.Icon})
	}
	return results
}

func ToMasterItemsResponseRelationshipTypes(types []services.RelationshipTypeDTO) []response.MasterItemResponse {
	var results []response.MasterItemResponse
	for _, t := range types {
		results = append(results, response.MasterItemResponse{ID: t.ID, Name: t.Name, Icon: t.Icon})
	}
	return results
}

func ToMasterItemsResponseInterests(interests []services.InterestDTO) []response.MasterItemResponse {
	var results []response.MasterItemResponse
	for _, i := range interests {
		results = append(results, response.MasterItemResponse{ID: i.ID, Name: i.Name, Icon: i.Icon})
	}
	return results
}

func ToMasterItemsResponseLanguages(languages []services.LanguageDTO) []response.MasterItemResponse {
	var results []response.MasterItemResponse
	for _, l := range languages {
		results = append(results, response.MasterItemResponse{ID: l.ID, Name: l.Name, Icon: l.Icon})
	}
	return results
}
