package monetization

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/delivery/http/response"
	"github.com/pipigendut/dating-backend/internal/repository"
	"github.com/pipigendut/dating-backend/internal/services"
)

type MonetizationHandler struct {
	subService services.SubscriptionService
	userRepo   repository.UserRepository
	storageService  *services.StorageService
}

func NewMonetizationHandler(r *gin.RouterGroup, subService services.SubscriptionService, userRepo repository.UserRepository, storageService *services.StorageService, authMiddleware gin.HandlerFunc) {
	handler := &MonetizationHandler{
		subService: subService,
		userRepo:   userRepo,
		storageService:  storageService,
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

	boostGroup := r.Group("/boosts")
	boostGroup.Use(authMiddleware)
	{
		boostGroup.GET("/availability", handler.GetBoostAvailability)
		boostGroup.POST("/activate", handler.ActivateBoost)
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
	response.OK(c, response.ToUserResponse(updatedUser, h.storageService))
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
	response.OK(c, response.ToUserResponse(updatedUser, h.storageService))
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

// GetBoostAvailability godoc
// @Summary Check if user has boosts available and current boost status
// @Tags Monetization
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.BaseResponse
// @Router /boosts/availability [get]
func (h *MonetizationHandler) GetBoostAvailability(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}
	userID := val.(uuid.UUID)
 
	entityID, err := uuid.Parse(c.Query("entity_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid entity_id parameter", err.Error())
		return
	}

	status, err := h.subService.GetStatus(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get boost availability", err.Error())
		return
	}

	isBoosted, expiresAt, err := h.subService.IsBoosted(c.Request.Context(), entityID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to check boost status", err.Error())
		return
	}

	boostAmount := status.Consumables["boost"]
	response.OK(c, gin.H{
		"has_boost":    boostAmount > 0,
		"boost_amount": boostAmount,
		"is_boosted":   isBoosted,
		"expires_at":   expiresAt,
	})
}

// ActivateBoost godoc
// @Summary Activate a boost for the user
// @Tags Monetization
// @Produce json
// @Security BearerAuth
// @Param request body ActivateBoostRequest true "Boost activation details"
// @Success 200 {object} response.BaseResponse
// @Router /boosts/activate [post]
func (h *MonetizationHandler) ActivateBoost(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}
	userID := val.(uuid.UUID)
 
	var req ActivateBoostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}
 
	boost, err := h.subService.ActivateBoost(c.Request.Context(), userID, req.EntityID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Failed to activate boost", err.Error())
		return
	}

	response.OK(c, gin.H{
		"message":    "Boost activated successfully",
		"expires_at": boost.ExpiresAt,
	})
}
type ActivateBoostRequest struct {
	EntityID uuid.UUID `json:"entity_id" binding:"required"`
}
