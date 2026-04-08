package v1

import "github.com/gin-gonic/gin"

func registerDeviceRoutes(v1 *gin.RouterGroup, h *Handlers, cfg RouterConfig) {
	devices := v1.Group("/devices")
	devices.Use(cfg.AuthMiddleware)
	{
		devices.POST("/register", h.Device.RegisterDevice)
		devices.PATCH("/fcm-token", h.Device.UpdateFCMToken)
		devices.POST("/deactivate", h.Device.DeactivateDevice)
	}
}
