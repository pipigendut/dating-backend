package v1

import "github.com/gin-gonic/gin"

func registerUserRoutes(v1 *gin.RouterGroup, h *Handlers, cfg RouterConfig) {
	// Public user routes
	v1.GET("/users/upload-url/public", h.User.GetUploadURLPublic)

	// Protected user routes
	users := v1.Group("/users")
	users.Use(cfg.AuthMiddleware)
	{
		users.GET("/profile/:id", h.User.GetProfile)
		users.PATCH("/profile", h.User.UpdateProfile)
		users.DELETE("/profile", h.User.DeleteAccount)
		users.GET("/upload-url", h.User.GetUploadURL)
		users.POST("/verify-face", h.User.VerifyFace)

		// Notifications
		notif := users.Group("/notifications")
		{
			notif.GET("", h.Notification.GetUserSettings)
			notif.POST("", h.Notification.UpdateUserSetting)
		}
	}
}
