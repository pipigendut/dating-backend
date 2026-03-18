package monetization

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/delivery/http/response"
	"github.com/pipigendut/dating-backend/internal/services"
)

type MonetizationHandler struct {
	subService services.SubscriptionService
}

func NewMonetizationHandler(r *gin.RouterGroup, subService services.SubscriptionService) {
	handler := &MonetizationHandler{
		subService: subService,
	}

	monGroup := r.Group("/monetization")
	{
		monGroup.GET("/plans", handler.GetPlans)
		monGroup.GET("/consumables", handler.GetConsumableItems)
		monGroup.POST("/purchase/consumable", handler.PurchaseConsumable)
		monGroup.POST("/purchase/plan", handler.PurchasePlan)
	}
}

// GetPlans godoc
// @Summary List all subscription plans
// @Description Get a list of all available subscription plans with prices and features
// @Tags Monetization
// @Produce json
// @Success 200 {object} response.BaseResponse
// @Router /monetization/plans [get]
func (h *MonetizationHandler) GetPlans(c *gin.Context) {
	plans, err := h.subService.GetPlans(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get plans", err.Error())
		return
	}
	response.OK(c, plans)
}

// GetConsumableItems godoc
// @Summary List all consumable items
// @Description Get a list of all buyable consumable packets (boosts, crushes)
// @Tags Monetization
// @Produce json
// @Success 200 {object} response.BaseResponse
// @Router /monetization/consumables [get]
func (h *MonetizationHandler) GetConsumableItems(c *gin.Context) {
	items, err := h.subService.GetConsumableItems(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get consumables", err.Error())
		return
	}
	response.OK(c, items)
}

type PurchaseRequest struct {
	ItemID  string `json:"item_id"`
	PlanID  string `json:"plan_id"`
	PriceID string `json:"price_id"`
}

// PurchaseConsumable godoc
// @Summary Purchase a consumable item
// @Tags Monetization
// @Accept json
// @Produce json
// @Param request body PurchaseRequest true "Purchase Request"
// @Success 200 {object} response.BaseResponse
// @Router /monetization/purchase/consumable [post]
func (h *MonetizationHandler) PurchaseConsumable(c *gin.Context) {
	var req PurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	err := h.subService.PurchaseConsumable(c.Request.Context(), userID.(uuid.UUID), uuid.MustParse(req.ItemID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to purchase consumable", err.Error())
		return
	}
	response.OK(c, "Purchase successful")
}

// PurchasePlan godoc
// @Summary Purchase a subscription plan
// @Tags Monetization
// @Accept json
// @Produce json
// @Param request body PurchaseRequest true "Purchase Request"
// @Success 200 {object} response.BaseResponse
// @Router /monetization/purchase/plan [post]
func (h *MonetizationHandler) PurchasePlan(c *gin.Context) {
	var req PurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	err := h.subService.PurchasePlan(c.Request.Context(), userID.(uuid.UUID), uuid.MustParse(req.PlanID), uuid.MustParse(req.PriceID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to purchase plan", err.Error())
		return
	}
	response.OK(c, "Subscription successful")
}
