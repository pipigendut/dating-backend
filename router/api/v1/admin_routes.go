package v1

import "github.com/gin-gonic/gin"

func registerAdminRoutes(v1 *gin.RouterGroup, h *Handlers, cfg RouterConfig) {
	admin := v1.Group("/admin")
	admin.Use(cfg.AuthMiddleware)
	admin.Use(h.Admin.WhitelistMiddleware())
	{
		configs := admin.Group("/configs")
		{
			configs.GET("", h.Admin.GetAllConfigs)
			configs.POST("/reload", h.Admin.ReloadConfigs)
			configs.POST("/reset", h.Admin.ResetConfigs)
		}

		admin.POST("/subscribe", h.Admin.SubscribeUser)
		consumables := admin.Group("/consumables")
		{
			consumables.POST("/boost", h.Admin.AddBoost)
			consumables.POST("/crush", h.Admin.AddCrush)
		}
	}
}
