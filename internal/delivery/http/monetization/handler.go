package monetization

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/delivery/http/response"
	userDTO "github.com/pipigendut/dating-backend/internal/delivery/http/user"
	"github.com/pipigendut/dating-backend/internal/repository"
	"github.com/pipigendut/dating-backend/internal/services"
)

type MonetizationHandler struct {
	subService services.SubscriptionService
	userRepo   repository.UserRepository
}

func NewMonetizationHandler(r *gin.RouterGroup, subService services.SubscriptionService, userRepo repository.UserRepository, authMiddleware gin.HandlerFunc) {
	handler := &MonetizationHandler{
		subService: subService,
		userRepo:   userRepo,
	}

	monGroup := r.Group("/monetization")
	monGroup.Use(authMiddleware)
	{
		monGroup.GET("/plans", handler.GetPlans)
		monGroup.GET("/consumables", handler.GetConsumableItems)
		monGroup.GET("/status", handler.GetStatus)
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
	updatedUser, err := h.userRepo.GetWithRelations(userID.(uuid.UUID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get updated user", err.Error())
		return
	}
	response.OK(c, userDTO.ToUserResponse(updatedUser))
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
	updatedUser, err := h.userRepo.GetWithRelations(userID.(uuid.UUID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get updated user", err.Error())
		return
	}
	response.OK(c, userDTO.ToUserResponse(updatedUser))
}

// GetStatus godoc
// @Summary Get current user's subscription status and features
// @Tags Monetization
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.BaseResponse
// @Router /monetization/status [get]
func (h *MonetizationHandler) GetStatus(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		// Fallback for some routers that might use different key
		val, exists = c.Get("user_id")
	}
	
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", "")
		return
	}

	userID := val.(uuid.UUID)
	status, err := h.subService.GetStatus(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get subscription status", err.Error())
		return
	}
	response.OK(c, status)
}
