package v1

import "github.com/gin-gonic/gin"

func registerMonetizationRoutes(v1 *gin.RouterGroup, h *Handlers, cfg RouterConfig) {
	mon := v1.Group("/monetization")
	mon.Use(cfg.AuthMiddleware)
	{
		mon.GET("/plans", h.Monetization.GetPlans)
		mon.GET("/consumables", h.Monetization.GetConsumableItems)
		mon.GET("/status", h.Monetization.GetStatus)
		mon.POST("/purchase/consumable", h.Monetization.PurchaseConsumable)
		mon.POST("/purchase/plan", h.Monetization.PurchasePlan)
	}

	boosts := v1.Group("/boosts")
	boosts.Use(cfg.AuthMiddleware)
	{
		boosts.GET("/availability", h.Monetization.GetBoostAvailability)
		boosts.POST("/activate", h.Monetization.ActivateBoost)
	}
}
