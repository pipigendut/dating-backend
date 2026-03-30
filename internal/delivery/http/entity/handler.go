package entity

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/delivery/http/response"
	"github.com/pipigendut/dating-backend/internal/entities"
	appErrors "github.com/pipigendut/dating-backend/internal/infra/errors"
	"github.com/pipigendut/dating-backend/internal/services"
)

type storageUsecase interface {
	GetPublicURL(key string) string
}

func NewEntityHandler(r *gin.RouterGroup, entitySvc services.EntityService, storageSvc storageUsecase, authMiddleware gin.HandlerFunc) {
	handler := &EntityHandler{
		entityService:  entitySvc,
		storageService: storageSvc,
	}

	entityGroup := r.Group("/entities")
	entityGroup.Use(authMiddleware)
	{
		entityGroup.GET("/:id", handler.GetEntity)
	}
}

type EntityHandler struct {
	entityService  services.EntityService
	storageService storageUsecase
}

// GetEntity godoc
// @Summary      Get entity details
// @Description  Fetches the complete profile details of an entity (solo user or group) by ID.
// @Tags         entities
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Entity ID"
// @Success      200  {object}  response.BaseResponse{data=response.EntityResponse} "Entity profile"
// @Failure      400  {object}  response.BaseResponse "Invalid request"
// @Failure      404  {object}  response.BaseResponse "Entity not found"
// @Failure      500  {object}  response.BaseResponse "Internal server error"
// @Router       /entities/{id} [get]
func (h *EntityHandler) GetEntity(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid entity ID", err.Error())
		return
	}

	ent, err := h.entityService.GetEntityByID(c.Request.Context(), id)
	if err != nil {
		if appErr, ok := err.(*appErrors.AppError); ok {
			response.Error(c, appErr.Code, appErr.Message, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Internal server error", err.Error())
		return
	}

	// Map to EntityResponse
	resp := response.EntityResponse{
		ID:   ent.ID,
		Type: string(ent.Type),
	}

	if ent.Type == entities.EntityTypeUser && ent.User != nil {
		ur := response.ToUserResponse(ent.User, h.storageService)
		resp.User = &ur
	} else if ent.Type == entities.EntityTypeGroup && ent.Group != nil {
		gr := response.ToGroupResponse(ent.Group, h.storageService)
		resp.Group = &gr
	}

	response.OK(c, resp)
}
