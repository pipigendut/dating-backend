package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pipigendut/dating-backend/internal/infra/errors"
	"github.com/pipigendut/dating-backend/internal/usecases"
)

type UserHandler struct {
	usecase *usecases.UserUsecase
}

func NewUserHandler(r *gin.RouterGroup, usecase *usecases.UserUsecase) {
	handler := &UserHandler{usecase: usecase}
	group := r.Group("/users")
	{
		group.GET("/profile/:id", handler.GetProfile)
	}
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	id := c.Param("id")
	user, err := h.usecase.GetProfile(id)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.Code, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, ToUserResponse(user))
}
