package auth

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/delivery/http/middleware"
	"github.com/pipigendut/dating-backend/internal/delivery/http/response"
	userHandler "github.com/pipigendut/dating-backend/internal/delivery/http/user"
	"github.com/pipigendut/dating-backend/internal/usecases"
)

type AuthHandler struct {
	usecase *usecases.AuthUsecase
}

func NewAuthHandler(r *gin.RouterGroup, usecase *usecases.AuthUsecase) {
	handler := &AuthHandler{usecase: usecase}
	group := r.Group("/auth")
	{
		group.POST("/check-email", handler.CheckEmail)
		group.POST("/register", handler.Register)
		group.POST("/login", handler.Login)
		group.POST("/google", handler.GoogleLogin)
		group.POST("/refresh", handler.Refresh)

		// Protected routes
		protected := group.Group("")
		protected.Use(middleware.AuthMiddleware())
		protected.POST("/logout", handler.Logout)
	}
}

// CheckEmail godoc
// @Summary      Check if email exists
// @Description  Checks if the provided email is already registered in the system.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      CheckEmailRequest  true  "Email to check"
// @Success      200      {object}  response.BaseResponse{data=CheckEmailResponse}
// @Failure      400      {object}  response.BaseResponse
// @Router       /auth/check-email [post]
func (h *AuthHandler) CheckEmail(c *gin.Context) {
	var req CheckEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	exists, err := h.usecase.CheckEmail(req.Email)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Internal server error", err.Error())
		return
	}

	response.OK(c, CheckEmailResponse{Exists: exists})
}

// Register godoc
// @Summary      Register via Email
// @Description  Creates a new user account using email and password.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      RegisterEmailRequest  true  "Registration details"
// @Success      200      {object}  response.BaseResponse{data=AuthResponse}
// @Failure      400      {object}  response.BaseResponse
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	var photosDTO *[]usecases.PhotoDTO
	if req.Photos != nil {
		mapped := make([]usecases.PhotoDTO, len(*req.Photos))
		for i, p := range *req.Photos {
			mapped[i] = usecases.PhotoDTO{URL: p.URL, IsMain: p.IsMain}
		}
		photosDTO = &mapped
	}

	dto := usecases.RegisterEmailDTO{
		ID:              req.ID,
		Email:           req.Email,
		Password:        req.Password,
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
		Device: usecases.DeviceDTO{
			DeviceID:    req.Device.DeviceID,
			DeviceName:  req.Device.DeviceName,
			DeviceModel: req.Device.DeviceModel,
			OSVersion:   req.Device.OSVersion,
			AppVersion:  req.Device.AppVersion,
			FCMToken:    req.Device.FCMToken,
			LastIP:      c.ClientIP(),
		},
	}

	token, refresh, user, err := h.usecase.RegisterEmail(dto)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	response.OK(c, AuthResponse{Token: token, RefreshToken: refresh, User: userHandler.ToUserResponse(user)})
}

// Login godoc
// @Summary      Login via Email
// @Description  Authenticates a user with email and password and returns a JWT token.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      LoginEmailRequest  true  "Login credentials"
// @Success      200      {object}  response.BaseResponse{data=AuthResponse}
// @Failure      401      {object}  response.BaseResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	dto := usecases.LoginEmailDTO{
		Email:    req.Email,
		Password: req.Password,
		Device: usecases.DeviceDTO{
			DeviceID:    req.Device.DeviceID,
			DeviceName:  req.Device.DeviceName,
			DeviceModel: req.Device.DeviceModel,
			OSVersion:   req.Device.OSVersion,
			AppVersion:  req.Device.AppVersion,
			FCMToken:    req.Device.FCMToken,
			LastIP:      c.ClientIP(),
		},
	}

	token, refresh, user, err := h.usecase.LoginEmail(dto)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	response.OK(c, AuthResponse{Token: token, RefreshToken: refresh, User: userHandler.ToUserResponse(user)})
}

// GoogleLogin godoc
// @Summary      Google OAuth Login/Register
// @Description  Handles Google OAuth authentication, including automatic account linking if the email already exists.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      GoogleLoginRequest  true  "Google OAuth data"
// @Success      200      {object}  response.BaseResponse{data=AuthResponse}
// @Failure      500      {object}  response.BaseResponse
// @Router       /auth/google [post]
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	var req GoogleLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	var photosDTO *[]usecases.PhotoDTO
	if req.Photos != nil {
		mapped := make([]usecases.PhotoDTO, len(*req.Photos))
		for i, p := range *req.Photos {
			mapped[i] = usecases.PhotoDTO{URL: p.URL, IsMain: p.IsMain}
		}
		photosDTO = &mapped
	}

	dto := usecases.GoogleLoginDTO{
		ID:              req.ID,
		Email:           req.Email,
		GoogleID:        req.GoogleID,
		FullName:        req.FullName,
		ProfilePicture:  req.ProfilePicture,
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
		Device: usecases.DeviceDTO{
			DeviceID:    req.Device.DeviceID,
			DeviceName:  req.Device.DeviceName,
			DeviceModel: req.Device.DeviceModel,
			OSVersion:   req.Device.OSVersion,
			AppVersion:  req.Device.AppVersion,
			FCMToken:    req.Device.FCMToken,
			LastIP:      c.ClientIP(),
		},
	}

	token, refresh, user, err := h.usecase.LoginWithGoogle(dto)
	if err != nil {
		log.Printf("[GoogleLogin Error] %v", err)
		response.Error(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	response.OK(c, AuthResponse{Token: token, RefreshToken: refresh, User: userHandler.ToUserResponse(user)})
}

// Refresh godoc
// @Summary      Refresh Access Token
// @Description  Allows the client to get a new access token using a valid refresh token and device ID.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      RefreshTokenRequest  true  "Refresh credentials"
// @Success      200      {object}  response.BaseResponse{data=TokenResponse}
// @Failure      401      {object}  response.BaseResponse
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	token, refresh, err := h.usecase.RefreshToken(req.RefreshToken, req.DeviceID)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	response.OK(c, TokenResponse{Token: token, RefreshToken: refresh})
}

// Logout godoc
// @Summary      Logout
// @Description  Revokes tokens and deactivates the associated device for the currently authenticated user.
// @Tags         auth
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      LogoutRequest  true  "Device data"
// @Success      200      {object}  response.BaseResponse
// @Failure      401      {object}  response.BaseResponse
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var uid uuid.UUID
	switch v := userID.(type) {
	case string:
		var err error
		uid, err = uuid.Parse(v)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid User ID format", nil)
			return
		}
	case uuid.UUID:
		uid = v
	default:
		response.Error(c, http.StatusBadRequest, "Invalid User ID type", nil)
		return
	}

	err := h.usecase.Logout(req.DeviceID, uid)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	response.OK(c, nil)
}
