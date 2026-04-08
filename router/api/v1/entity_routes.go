package v1

import "github.com/gin-gonic/gin"

func registerEntityRoutes(v1 *gin.RouterGroup, h *Handlers, cfg RouterConfig) {
	v1.GET("/entities/:id", cfg.AuthMiddleware, h.Entity.GetEntity)
}
