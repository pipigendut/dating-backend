package v1

import "github.com/gin-gonic/gin"

func registerMasterRoutes(v1 *gin.RouterGroup, h *Handlers, cfg RouterConfig) {
	masterGroup := v1.Group("/master")
	masterGroup.Use(cfg.BasicAuthMiddleware)
	{
		masterGroup.GET("/genders", h.Master.GetGenders)
		masterGroup.GET("/relationship-types", h.Master.GetRelationshipTypes)
		masterGroup.GET("/interests", h.Master.GetInterests)
		masterGroup.GET("/languages", h.Master.GetLanguages)
	}

	// Advertisements now require token login (JWT)
	v1.GET("/advertisements", cfg.AuthMiddleware, h.Master.GetAdvertisements)
}
