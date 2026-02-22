package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pipigendut/dating-backend/internal/delivery/http/response"
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

	token, err := h.usecase.RegisterEmail(req.Email, req.Password)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	response.OK(c, AuthResponse{Token: token})
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

	token, err := h.usecase.LoginEmail(req.Email, req.Password)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	response.OK(c, AuthResponse{Token: token})
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

	dto := usecases.GoogleLoginDTO{
		Email:          req.Email,
		GoogleID:       req.GoogleID,
		FullName:       req.FullName,
		ProfilePicture: req.ProfilePicture,
	}

	token, err := h.usecase.LoginWithGoogle(dto)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	response.OK(c, AuthResponse{Token: token})
}
