package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	base "github.com/pipigendut/dating-backend/internal/delivery/http/dto"
	dtov1 "github.com/pipigendut/dating-backend/internal/delivery/http/dto/v1"
	"github.com/pipigendut/dating-backend/internal/delivery/http/mapper"
	"github.com/pipigendut/dating-backend/internal/entities"
	appErrors "github.com/pipigendut/dating-backend/internal/infra/errors"
	"github.com/pipigendut/dating-backend/internal/services"
)

type storageUsecase interface {
	GetPublicURL(key string) string
}

func NewEntityHandler(entitySvc services.EntityService, storageSvc storageUsecase) *EntityHandler {
	return &EntityHandler{
		entityService:  entitySvc,
		storageService: storageSvc,
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
// @Success      200  {object}  base.BaseResponse{data=dtov1.EntityResponse} "Entity profile"
// @Failure      400  {object}  base.BaseResponse "Invalid request"
// @Failure      404  {object}  base.BaseResponse "Entity not found"
// @Failure      500  {object}  base.BaseResponse "Internal server error"
// @Router       /entities/{id} [get]
func (h *EntityHandler) GetEntity(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		base.Error(c, http.StatusBadRequest, "Invalid entity ID", err.Error())
		return
	}

	ent, err := h.entityService.GetEntityByID(c.Request.Context(), id)
	if err != nil {
		if appErr, ok := err.(*appErrors.AppError); ok {
			base.Error(c, appErr.Code, appErr.Message, nil)
			return
		}
		base.Error(c, http.StatusInternalServerError, "Internal server error", err.Error())
		return
	}

	// Map to EntityResponse
	resp := dtov1.EntityResponse{
		ID:   ent.ID,
		Type: string(ent.Type),
	}

	if ent.Type == entities.EntityTypeUser && ent.User != nil {
		ur := mapper.ToUserResponse(ent.User, h.storageService)
		resp.User = &ur
	} else if ent.Type == entities.EntityTypeGroup && ent.Group != nil {
		gr := mapper.ToGroupResponse(ent.Group, h.storageService)
		resp.Group = &gr
	}

	base.OK(c, resp)
}
