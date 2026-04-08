package v1

import "github.com/gin-gonic/gin"

func registerAuthRoutes(v1 *gin.RouterGroup, h *Handlers, cfg RouterConfig) {
	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/check-email", h.Auth.CheckEmail)
		authGroup.POST("/register", h.Auth.Register)
		authGroup.POST("/login", h.Auth.Login)
		authGroup.POST("/google", h.Auth.GoogleLogin)
		authGroup.POST("/refresh", h.Auth.Refresh)
		
		// Logout requires auth
		authGroup.POST("/logout", cfg.AuthMiddleware, h.Auth.Logout)
	}
}
