package v1

import "github.com/gin-gonic/gin"

func registerSwipeRoutes(v1 *gin.RouterGroup, h *Handlers, cfg RouterConfig) {
	swipe := v1.Group("/swipe")
	swipe.Use(cfg.AuthMiddleware)
	{
		swipe.GET("/candidates", h.Swipe.GetCandidates)
		swipe.POST("/", h.Swipe.Swipe)
		swipe.GET("/likes", h.Swipe.GetIncomingLikes)
		swipe.GET("/likes/sent", h.Swipe.GetLikesSent)
		swipe.POST("/unmatch/:entity_id", h.Swipe.Unmatch)
		swipe.DELETE("/unlike/:entity_id", h.Swipe.Unlike)
		swipe.GET("/likes/count", h.Swipe.GetLikesCount)
	}
}
