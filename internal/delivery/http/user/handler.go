package user

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/delivery/http/response"
	"github.com/pipigendut/dating-backend/internal/entities"
	appErrors "github.com/pipigendut/dating-backend/internal/infra/errors"
	"github.com/pipigendut/dating-backend/internal/services"
)

func NewUserHandler(r *gin.RouterGroup, userService *services.UserService, storageService *services.StorageService, verifyService *services.VerificationService, entitySvc services.EntityService, authMiddleware gin.HandlerFunc) {
	handler := &UserHandler{
		userService:    userService,
		storageService: storageService,
		verifyService:  verifyService,
		entityService:  entitySvc,
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
	userService    *services.UserService
	storageService *services.StorageService
	verifyService  *services.VerificationService
	entityService  services.EntityService
}

// GetProfile godoc
// @Summary      Get user profile
// @Description  Fetches the profile details of a specific user by ID.
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  response.BaseResponse{data=response.UserResponse} "User profile"
// @Failure      400  {object}  response.BaseResponse "Invalid request"
// @Failure      500  {object}  response.BaseResponse "Internal server error"
// @Router       /users/profile/{id} [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	id := c.Param("id")
	user, err := h.userService.GetProfile(id)
	if err != nil {
		if appErr, ok := err.(*appErrors.AppError); ok {
			response.Error(c, appErr.Code, appErr.Message, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Internal server error", err.Error())
		return
	}

	response.OK(c, response.ToUserResponse(user, h.storageService))
}

// ... (Update other methods similarly in my head or just use sed later, but I'll do a few major ones here)

// UpdateProfile godoc
// @Summary      Update user profile
// @Description  Updates the authenticated user's profile information, including photos and preferences.
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body UpdateProfileRequest true "Profile update details"
// @Success      200  {object}  response.BaseResponse{data=response.UserResponse} "Updated profile"
// @Router       /users/profile [patch]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	var photosDTO *[]services.PhotoDTO
	if req.Photos != nil {
		mapped := make([]services.PhotoDTO, len(*req.Photos))
		for i, p := range *req.Photos {
			mapped[i] = services.PhotoDTO{ID: p.ID, URL: p.URL, IsMain: p.IsMain, Destroy: p.Destroy}
		}
		photosDTO = &mapped
	}

	serviceReq := services.UpdateProfileRequest{
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
		serviceReq.Status = &status
	}

	if err := h.userService.UpdateProfile(userID, serviceReq); err != nil {
		if appErr, ok := err.(*appErrors.AppError); ok {
			response.Error(c, appErr.Code, appErr.Message, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to update profile", err.Error())
		return
	}

	updatedUser, err := h.userService.GetProfile(userID.String())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch updated profile", err.Error())
		return
	}

	response.OK(c, response.ToUserResponse(updatedUser, h.storageService))
}

// DeleteAccount godoc
// @Summary      Delete user account
// @Description  Permanently deletes the authenticated user's account and all associated data.
// @Tags         users
// @Security     BearerAuth
// @Success      200  {object}  response.BaseResponse "Successfully deleted"
// @Router       /users/profile [delete]
func (h *UserHandler) DeleteAccount(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	if err := h.userService.DeleteAccount(userID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete account", err.Error())
		return
	}

	response.OK(c, nil)
}

// GetUploadURL godoc
// @Summary      Get private upload URL
// @Description  Generates a temporary S3 upload URL for private user media.
// @Tags         users
// @Security     BearerAuth
// @Success      200  {object}  response.BaseResponse{data=UploadURLResponse} "Upload URL details"
// @Router       /users/upload-url [get]
func (h *UserHandler) GetUploadURL(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	url, key, err := h.storageService.GetUploadURL(c.Request.Context(), userID)
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
// @Summary      Get public upload URL
// @Description  Generates a temporary S3 upload URL for publicly accessible assets (e.g., registration photos).
// @Tags         users
// @Param        client_id query string true "Client Identifier"
// @Success      200  {object}  response.BaseResponse{data=UploadURLResponse} "Upload URL details"
// @Router       /users/upload-url/public [get]
func (h *UserHandler) GetUploadURLPublic(c *gin.Context) {
	clientID := c.Query("client_id")
	if clientID == "" {
		response.Error(c, http.StatusBadRequest, "client_id is required", nil)
		return
	}

	url, key, err := h.storageService.GetUploadURLPublic(c.Request.Context(), clientID)
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
// @Description  Performs facial verification by comparing an uploaded snapshot against the user's profile photo.
// @Tags         users
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        photo formData file true "Face snapshot"
// @Success      200  {object}  response.BaseResponse{data=response.VerificationResult} "Verification result"
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

	result, err := h.verifyService.VerifyFace(c.Request.Context(), userID, f)
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


func (h *UserHandler) parseUUIDs(strs []string) []uuid.UUID {
	var uuids []uuid.UUID
	for _, s := range strs {
		if u, err := uuid.Parse(s); err == nil {
			uuids = append(uuids, u)
		}
	}
	return uuids
}




