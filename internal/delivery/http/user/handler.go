package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pipigendut/dating-backend/internal/delivery/http/response"
	"github.com/pipigendut/dating-backend/internal/infra/errors"
	"github.com/pipigendut/dating-backend/internal/usecases"
)

type UserHandler struct {
	usecase *usecases.UserUsecase
}

func NewUserHandler(r *gin.RouterGroup, usecase *usecases.UserUsecase) {
	handler := &UserHandler{usecase: usecase}
	users := r.Group("/users")
	{
		users.GET("/profile/:id", handler.GetProfile)
	}
}

// GetProfile godoc
// @Summary      Get user profile
// @Description  Retrieves the full profile of a user by their UUID.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "User ID (UUID)"
// @Success      200  {object}  response.BaseResponse{data=UserResponse}
// @Failure      404  {object}  response.BaseResponse
// @Security     BearerAuth
// @Router       /users/profile/{id} [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	id := c.Param("id")
	user, err := h.usecase.GetProfile(id)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.Error(c, appErr.Code, appErr.Message, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Internal server error", err.Error())
		return
	}

	response.OK(c, ToUserResponse(user))
}
