package master

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pipigendut/dating-backend/internal/delivery/http/response"
	"github.com/pipigendut/dating-backend/internal/usecases"
)

type MasterHandler struct {
	usecase *usecases.MasterUsecase
}

func NewMasterHandler(r *gin.RouterGroup, usecase *usecases.MasterUsecase) {
	handler := &MasterHandler{usecase: usecase}
	group := r.Group("/master")
	{
		group.GET("/genders", handler.GetGenders)
		group.GET("/relationship-types", handler.GetRelationshipTypes)
		group.GET("/interests", handler.GetInterests)
		group.GET("/languages", handler.GetLanguages)
	}
}

// GetGenders godoc
// @Summary      Get all active genders
// @Description  Returns a list of all active gender options for user profiling.
// @Tags         master
// @Produce      json
// @Success      200      {object}  response.BaseResponse{data=[]MasterItemResponse}
// @Failure      500      {object}  response.BaseResponse
// @Router       /master/genders [get]
func (h *MasterHandler) GetGenders(c *gin.Context) {
	items, err := h.usecase.GetGenders()
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
// @Success      200      {object}  response.BaseResponse{data=[]MasterItemResponse}
// @Failure      500      {object}  response.BaseResponse
// @Router       /master/relationship-types [get]
func (h *MasterHandler) GetRelationshipTypes(c *gin.Context) {
	items, err := h.usecase.GetRelationshipTypes()
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
// @Success      200      {object}  response.BaseResponse{data=[]MasterItemResponse}
// @Failure      500      {object}  response.BaseResponse
// @Router       /master/interests [get]
func (h *MasterHandler) GetInterests(c *gin.Context) {
	items, err := h.usecase.GetInterests()
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
// @Success      200      {object}  response.BaseResponse{data=[]MasterItemResponse}
// @Failure      500      {object}  response.BaseResponse
// @Router       /master/languages [get]
func (h *MasterHandler) GetLanguages(c *gin.Context) {
	items, err := h.usecase.GetLanguages()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve languages", err.Error())
		return
	}
	response.OK(c, ToMasterItemsResponseLanguages(items))
}
