package v1

import "github.com/gin-gonic/gin"

func registerGroupRoutes(v1 *gin.RouterGroup, h *Handlers, cfg RouterConfig) {
	// Public group invite validation
	v1.GET("/group-invites/validate", h.Group.ValidateInvite)

	groups := v1.Group("/groups")
	groups.Use(cfg.AuthMiddleware)
	{
		groups.POST("", h.Group.CreateGroup)
		groups.GET("/me", h.Group.GetMyGroup)
		groups.DELETE("/:id", h.Group.DisbandGroup)
		groups.POST("/:id/leave", h.Group.LeaveGroup)
		groups.DELETE("/:id/members/:user_id", h.Group.KickMember)
		groups.POST("/:id/invite", h.Group.InviteToGroup)
		groups.POST("/:id/invite-link", h.Group.GenerateInviteLink)
	}

	// Protected accept invite
	v1.POST("/group-invites/accept", cfg.AuthMiddleware, h.Group.AcceptInvite)
}
