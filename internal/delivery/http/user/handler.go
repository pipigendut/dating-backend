package user

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/delivery/http/response"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/infra/errors"
	"github.com/pipigendut/dating-backend/internal/usecases"
)

func NewUserHandler(r *gin.RouterGroup, usecase *usecases.UserUsecase, storageUC *usecases.StorageUsecase, verifyUC *usecases.VerificationService, authMiddleware gin.HandlerFunc) {
	handler := &UserHandler{
		usecase:        usecase,
		storageUC:      storageUC,
		verificationUC: verifyUC,
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
		users.POST("/verify-face", handler.VerifyFace)
	}
}

type UserHandler struct {
	usecase        *usecases.UserUsecase
	storageUC      *usecases.StorageUsecase
	verificationUC *usecases.VerificationService
}

// GetProfile godoc
// @Summary      Get user profile
// @Description  Fetches the profile details of a specific user by ID.
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  response.BaseResponse{data=UserResponse} "User profile"
// @Failure      400  {object}  response.BaseResponse "Invalid request"
// @Failure      500  {object}  response.BaseResponse "Internal server error"
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

	response.OK(c, ToUserResponse(user, h.storageUC))
}

// UpdateProfile godoc
// @Summary      Update user profile
// @Description  Updates the current user's profile information.
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body UpdateProfileRequest true "Update profile details"
// @Success      200  {object}  response.BaseResponse{data=UserResponse} "Updated user profile"
// @Failure      400  {object}  response.BaseResponse "Invalid request"
// @Failure      500  {object}  response.BaseResponse "Internal server error"
// @Router       /users/profile [patch]
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
		FullName:         req.FullName,
		DateOfBirth:      req.DateOfBirth,
		Gender:           req.Gender,
		HeightCM:         req.HeightCM,
		Bio:              req.Bio,
		InterestedIn:     req.InterestedIn,
		RelationshipType: req.RelationshipType,
		LocationCity:     req.LocationCity,
		LocationCountry:  req.LocationCountry,
		Latitude:         req.Latitude,
		Longitude:        req.Longitude,
		Interests:        req.Interests,
		Languages:        req.Languages,
		Photos:           photosDTO,
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

	updatedUser, err := h.usecase.GetProfile(userID.String())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch updated profile", err.Error())
		return
	}

	response.OK(c, ToUserResponse(updatedUser, h.storageUC))
}

// DeleteAccount godoc
// @Summary      Delete user account
// @Description  Permanently deletes the current user's account and all associated data.
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.BaseResponse "Successfully deleted account"
// @Failure      500  {object}  response.BaseResponse "Internal server error"
// @Router       /users/profile [delete]
func (h *UserHandler) DeleteAccount(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	if err := h.usecase.DeleteAccount(userID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete account", err.Error())
		return
	}

	response.OK(c, nil)
}

// GetUploadURL godoc
// @Summary      Get presigned S3 upload URL
// @Description  Generates a presigned URL for the user to securely upload photos to S3.
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.BaseResponse{data=UploadURLResponse} "Upload details"
// @Failure      500  {object}  response.BaseResponse "Internal server error"
// @Router       /users/upload-url [get]
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

// GetUploadURLPublic godoc
// @Summary      Get public presigned S3 upload URL
// @Description  Generates a presigned URL for a new (unauthenticated) user during onboarding.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        client_id query string true "Client ID constraint"
// @Success      200  {object}  response.BaseResponse{data=UploadURLResponse} "Upload details"
// @Failure      400  {object}  response.BaseResponse "Invalid request"
// @Failure      500  {object}  response.BaseResponse "Internal server error"
// @Router       /users/upload-url/public [get]
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

// VerifyFace godoc
// @Summary      Verify user face
// @Description  Uploads a photo to perform facial verification heuristics for profile validation.
// @Tags         users
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        photo formData file true "Photo to verify"
// @Success      200  {object}  response.BaseResponse "Verification result"
// @Failure      400  {object}  response.BaseResponse "Invalid request"
// @Failure      429  {object}  response.BaseResponse "Too many requests"
// @Failure      500  {object}  response.BaseResponse "Internal server error"
// @Router       /users/verify-face [post]
func (h *UserHandler) VerifyFace(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	file, err := c.FormFile("photo")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Photo is required", err.Error())
		return
	}

	f, err := file.Open()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to open photo", err.Error())
		return
	}
	defer f.Close()

	result, err := h.verificationUC.VerifyFace(c.Request.Context(), userID, f)
	if err != nil {
		if strings.Contains(err.Error(), "daily face verification limit exceeded") {
			response.Error(c, http.StatusTooManyRequests, "Daily Limit Exceeded", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Verification process failed", err.Error())
		return
	}

	response.OK(c, result)
}
