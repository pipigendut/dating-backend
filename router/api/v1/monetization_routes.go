package v1

import "github.com/gin-gonic/gin"

func registerMonetizationRoutes(v1 *gin.RouterGroup, h *Handlers, cfg RouterConfig) {
	mon := v1.Group("/monetization")
	{
		// These endpoints use Basic Auth (Basic Token)
		mon.GET("/plans", cfg.BasicAuthMiddleware, h.Monetization.GetPlans)
		mon.GET("/consumables", cfg.BasicAuthMiddleware, h.Monetization.GetConsumableItems)

		// These endpoints require token login (JWT)
		monAuth := mon.Group("/")
		monAuth.Use(cfg.AuthMiddleware)
		{
			monAuth.GET("/status", h.Monetization.GetStatus)
			monAuth.POST("/purchase/consumable", h.Monetization.PurchaseConsumable)
			monAuth.POST("/purchase/plan", h.Monetization.PurchasePlan)
		}
	}

	boosts := v1.Group("/boosts")
	boosts.Use(cfg.AuthMiddleware)
	{
		boosts.GET("/availability", h.Monetization.GetBoostAvailability)
		boosts.POST("/activate", h.Monetization.ActivateBoost)
	}
}
