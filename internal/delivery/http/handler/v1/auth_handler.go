package v1

import (
	dtov1 "github.com/pipigendut/dating-backend/internal/delivery/http/dto/v1"

	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	base "github.com/pipigendut/dating-backend/internal/delivery/http/dto"
	"github.com/pipigendut/dating-backend/internal/delivery/http/mapper"
	"github.com/pipigendut/dating-backend/internal/services"
)

type AuthHandler struct {
	svc            *services.AuthService
	storageService *services.StorageService
}

func NewAuthHandler(svc *services.AuthService, storageService *services.StorageService) *AuthHandler {
	return &AuthHandler{svc: svc, storageService: storageService}
}

// CheckEmail godoc
// @Summary      Check if email exists
// @Description  Checks if the provided email is already registered in the system.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dtov1.CheckEmailRequest  true  "Email to check"
// @Success      200      {object}  base.BaseResponse{data=dtov1.CheckEmailResponse}
// @Failure      400      {object}  base.BaseResponse
// @Router       /auth/check-email [post]
func (h *AuthHandler) CheckEmail(c *gin.Context) {
	var req dtov1.CheckEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	exists, err := h.svc.CheckEmail(req.Email)
	if err != nil {
		base.Error(c, http.StatusInternalServerError, "Internal server error", err.Error())
		return
	}

	base.OK(c, dtov1.CheckEmailResponse{Exists: exists})
}

// Register godoc
// @Summary      Register via Email
// @Description  Creates a new user account using email and password.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dtov1.RegisterEmailRequest  true  "Registration details"
// @Success      200      {object}  base.BaseResponse{data=dtov1.AuthResponse}
// @Failure      400      {object}  base.BaseResponse
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dtov1.RegisterEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	var photosDTO *[]services.PhotoDTO
	if req.Photos != nil {
		mapped := make([]services.PhotoDTO, len(*req.Photos))
		for i, p := range *req.Photos {
			mapped[i] = services.PhotoDTO{URL: p.URL, IsMain: p.IsMain}
		}
		photosDTO = &mapped
	}

	dto := services.RegisterEmailDTO{
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
		Device: services.DeviceDTO{
			DeviceID:    req.Device.DeviceID,
			DeviceName:  req.Device.DeviceName,
			DeviceModel: req.Device.DeviceModel,
			OSVersion:   req.Device.OSVersion,
			AppVersion:  req.Device.AppVersion,
			FCMToken:    req.Device.FCMToken,
			LastIP:      c.ClientIP(),
		},
	}

	token, refresh, user, err := h.svc.RegisterEmail(dto)
	if err != nil {
		base.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	base.OK(c, dtov1.AuthResponse{Token: token, RefreshToken: refresh, User: mapper.ToUserResponse(user, h.storageService)})
}

// Login godoc
// @Summary      Login via Email
// @Description  Authenticates a user with email and password and returns a JWT token.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dtov1.LoginEmailRequest  true  "Login credentials"
// @Success      200      {object}  base.BaseResponse{data=dtov1.AuthResponse}
// @Failure      401      {object}  base.BaseResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dtov1.LoginEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	dto := services.LoginEmailDTO{
		Email:    req.Email,
		Password: req.Password,
		Device: services.DeviceDTO{
			DeviceID:    req.Device.DeviceID,
			DeviceName:  req.Device.DeviceName,
			DeviceModel: req.Device.DeviceModel,
			OSVersion:   req.Device.OSVersion,
			AppVersion:  req.Device.AppVersion,
			FCMToken:    req.Device.FCMToken,
			LastIP:      c.ClientIP(),
		},
	}

	token, refresh, user, err := h.svc.LoginEmail(dto)
	if err != nil {
		base.Error(c, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	base.OK(c, dtov1.AuthResponse{Token: token, RefreshToken: refresh, User: mapper.ToUserResponse(user, h.storageService)})
}

// GoogleLogin godoc
// @Summary      Google OAuth Login/Register
// @Description  Handles Google OAuth authentication, including automatic account linking if the email already exists.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dtov1.GoogleLoginRequest  true  "Google OAuth data"
// @Success      200      {object}  base.BaseResponse{data=dtov1.AuthResponse}
// @Failure      500      {object}  base.BaseResponse
// @Router       /auth/google [post]
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	var req dtov1.GoogleLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	var photosDTO *[]services.PhotoDTO
	if req.Photos != nil {
		mapped := make([]services.PhotoDTO, len(*req.Photos))
		for i, p := range *req.Photos {
			mapped[i] = services.PhotoDTO{URL: p.URL, IsMain: p.IsMain}
		}
		photosDTO = &mapped
	}

	dto := services.GoogleLoginDTO{
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
		Device: services.DeviceDTO{
			DeviceID:    req.Device.DeviceID,
			DeviceName:  req.Device.DeviceName,
			DeviceModel: req.Device.DeviceModel,
			OSVersion:   req.Device.OSVersion,
			AppVersion:  req.Device.AppVersion,
			FCMToken:    req.Device.FCMToken,
			LastIP:      c.ClientIP(),
		},
	}

	token, refresh, user, err := h.svc.LoginWithGoogle(dto)
	if err != nil {
		log.Printf("[GoogleLogin Error] %v", err)
		base.Error(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	base.OK(c, dtov1.AuthResponse{Token: token, RefreshToken: refresh, User: mapper.ToUserResponse(user, h.storageService)})
}

// Refresh godoc
// @Summary      Refresh Access Token
// @Description  Allows the client to get a new access token using a valid refresh token and device ID.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dtov1.RefreshTokenRequest  true  "Refresh credentials"
// @Success      200      {object}  base.BaseResponse{data=dtov1.TokenResponse}
// @Failure      401      {object}  base.BaseResponse
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dtov1.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	token, refresh, err := h.svc.RefreshToken(req.RefreshToken, req.DeviceID)
	if err != nil {
		base.Error(c, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	base.OK(c, dtov1.TokenResponse{Token: token, RefreshToken: refresh})
}

// Logout godoc
// @Summary      Logout
// @Description  Revokes tokens and deactivates the associated device for the currently authenticated user.
// @Tags         auth
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      dtov1.LogoutRequest  true  "Device data"
// @Success      200      {object}  base.BaseResponse
// @Failure      401      {object}  base.BaseResponse
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req dtov1.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		base.Error(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var uid uuid.UUID
	switch v := userID.(type) {
	case string:
		var err error
		uid, err = uuid.Parse(v)
		if err != nil {
			base.Error(c, http.StatusBadRequest, "Invalid User ID format", nil)
			return
		}
	case uuid.UUID:
		uid = v
	default:
		base.Error(c, http.StatusBadRequest, "Invalid User ID type", nil)
		return
	}

	err := h.svc.Logout(req.DeviceID, uid)
	if err != nil {
		base.Error(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	base.OK(c, nil)
}
