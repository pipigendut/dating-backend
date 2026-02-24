package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/delivery/http/response"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/infra/errors"
	"github.com/pipigendut/dating-backend/internal/usecases"
)

func NewUserHandler(r *gin.RouterGroup, usecase *usecases.UserUsecase, storageUC *usecases.StorageUsecase, authMiddleware gin.HandlerFunc) {
	handler := &UserHandler{
		usecase:   usecase,
		storageUC: storageUC,
	}
	users := r.Group("/users")

	// Unauthenticated endpoints
	users.GET("/upload-url/public", handler.GetUploadURLPublic)

	// Authenticated endpoints
	users.Use(authMiddleware)
	{
		users.GET("/profile/:id", handler.GetProfile)
		users.PATCH("/profile", handler.UpdateProfile)
		users.DELETE("/profile", handler.DeleteAccount)
		users.GET("/upload-url", handler.GetUploadURL)
	}
}

type UserHandler struct {
	usecase   *usecases.UserUsecase
	storageUC *usecases.StorageUsecase
}

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

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	var photosDTO *[]usecases.PhotoDTO
	if req.Photos != nil {
		mapped := make([]usecases.PhotoDTO, len(*req.Photos))
		for i, p := range *req.Photos {
			mapped[i] = usecases.PhotoDTO{ID: p.ID, URL: p.URL, IsMain: p.IsMain, Destroy: p.Destroy}
		}
		photosDTO = &mapped
	}

	usecaseReq := usecases.UpdateProfileRequest{
		FullName:        req.FullName,
		DateOfBirth:     req.DateOfBirth,
		Gender:          req.Gender,
		HeightCM:        req.HeightCM,
		Bio:             req.Bio,
		InterestedIn:    req.InterestedIn,
		LookingFor:      req.LookingFor,
		LocationCity:    req.LocationCity,
		LocationCountry: req.LocationCountry,
		Latitude:        req.Latitude,
		Longitude:       req.Longitude,
		Interests:       req.Interests,
		Languages:       req.Languages,
		Photos:          photosDTO,
	}

	if req.Status != nil {
		status := entities.UserStatus(*req.Status)
		usecaseReq.Status = &status
	}

	if err := h.usecase.UpdateProfile(userID, usecaseReq); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.Error(c, appErr.Code, appErr.Message, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to update profile", err.Error())
		return
	}

	response.OK(c, nil)
}

func (h *UserHandler) DeleteAccount(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	if err := h.usecase.DeleteAccount(userID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete account", err.Error())
		return
	}

	response.OK(c, nil)
}

func (h *UserHandler) GetUploadURL(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	url, key, err := h.storageUC.GetUploadURL(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to generate upload URL", err.Error())
		return
	}

	response.OK(c, UploadURLResponse{
		UploadURL: url,
		FileKey:   key,
	})
}

func (h *UserHandler) GetUploadURLPublic(c *gin.Context) {
	clientID := c.Query("client_id")
	if clientID == "" {
		response.Error(c, http.StatusBadRequest, "client_id is required", nil)
		return
	}

	url, key, err := h.storageUC.GetUploadURLPublic(c.Request.Context(), clientID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to generate public upload URL", err.Error())
		return
	}

	response.OK(c, UploadURLResponse{
		UploadURL: url,
		FileKey:   key,
	})
}
