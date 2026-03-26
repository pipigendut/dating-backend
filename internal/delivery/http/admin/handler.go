package admin

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/delivery/http/response"
	userDTO "github.com/pipigendut/dating-backend/internal/delivery/http/user"
	"github.com/pipigendut/dating-backend/internal/infra/seeds"
	"github.com/pipigendut/dating-backend/internal/repository"
	"github.com/pipigendut/dating-backend/internal/services"
	"github.com/pipigendut/dating-backend/internal/usecases"
	"gorm.io/gorm"
)

type AdminHandler struct {
	db        *gorm.DB
	configSvc services.ConfigService
	adminSvc  services.AdminService
	userRepo  repository.UserRepository
	storageUC *usecases.StorageUsecase
}

func NewAdminHandler(db *gorm.DB, configSvc services.ConfigService, adminSvc services.AdminService, userRepo repository.UserRepository, storageUC *usecases.StorageUsecase) *AdminHandler {
	return &AdminHandler{
		db:        db,
		configSvc: configSvc,
		adminSvc:  adminSvc,
		userRepo:  userRepo,
		storageUC: storageUC,
	}
}

func (h *AdminHandler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	admin := rg.Group("/admin")
	admin.Use(authMiddleware)
	admin.Use(h.WhitelistMiddleware())
	{
		configs := admin.Group("/configs")
		{
			configs.GET("", h.GetAllConfigs)
			configs.POST("/reload", h.ReloadConfigs)
			configs.POST("/reset", h.ResetConfigs)
		}

		admin.POST("/subscribe", h.SubscribeUser)
		consumables := admin.Group("/consumables")
		{
			consumables.POST("/boost", h.AddBoost)
			consumables.POST("/crush", h.AddCrush)
		}
	}
}

func (h *AdminHandler) WhitelistMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		whitelistStr := h.configSvc.GetString("whitelist_emails", "")
		if whitelistStr == "" {
			c.Next()
			return
		}

		val, exists := c.Get("userID")
		if !exists {
			val, exists = c.Get("user_id")
		}

		if !exists {
			response.Error(c, http.StatusUnauthorized, "User not identified", "")
			c.Abort()
			return
		}

		userID := val.(uuid.UUID)
		user, err := h.userRepo.GetByID(userID)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to verify whitelist", err.Error())
			c.Abort()
			return
		}

		var allowedEmails []string
		if strings.HasPrefix(whitelistStr, "[") {
			if err := json.Unmarshal([]byte(whitelistStr), &allowedEmails); err != nil {
				response.Error(c, http.StatusInternalServerError, "Failed to parse whitelist", err.Error())
				c.Abort()
				return
			}
		} else {
			allowedEmails = strings.Split(whitelistStr, ",")
		}

		isAllowed := false
		userEmail := ""
		if user != nil && user.Email != nil {
			userEmail = *user.Email
		}

		for _, email := range allowedEmails {
			if strings.TrimSpace(email) == userEmail {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			response.Error(c, http.StatusForbidden, "Email not whitelisted for admin actions", userEmail)
			c.Abort()
			return
		}

		c.Next()
	}
}

type AdminSubscribeRequest struct {
	UserID string `json:"user_id" binding:"required"`
	PlanID string `json:"plan_id" binding:"required"`
}

type AdminConsumableRequest struct {
	UserID    string `json:"user_id" binding:"required"`
	PackageID string `json:"package_id" binding:"required"`
}

// SubscribeUser godoc
// @Summary Simulate user subscription to a plan
// @Tags admin
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body AdminSubscribeRequest true "Subscribe Request"
// @Success 200 {object} userDTO.UserResponse
// @Router /admin/subscribe [post]
func (h *AdminHandler) SubscribeUser(c *gin.Context) {
	var req AdminSubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	uid, err := uuid.Parse(req.UserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user_id", err.Error())
		return
	}

	pid, err := uuid.Parse(req.PlanID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid plan_id", err.Error())
		return
	}

	user, err := h.adminSvc.SubscribeUser(c.Request.Context(), uid, pid)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to subscribe user", err.Error())
		return
	}

	response.OK(c, userDTO.ToUserResponse(user, h.storageUC))
}

// AddBoost godoc
// @Summary Simulate adding a boost package to user
// @Tags admin
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body AdminConsumableRequest true "Boost Request"
// @Success 200 {object} userDTO.UserResponse
// @Router /admin/consumables/boost [post]
func (h *AdminHandler) AddBoost(c *gin.Context) {
	var req AdminConsumableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	uid, err := uuid.Parse(req.UserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user_id", err.Error())
		return
	}

	pkgID, err := uuid.Parse(req.PackageID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid package_id", err.Error())
		return
	}

	// Note: ItemType check could be added if needed, but AdminService.AddConsumable handles any package.
	user, err := h.adminSvc.AddConsumable(c.Request.Context(), uid, pkgID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to add boost", err.Error())
		return
	}

	response.OK(c, userDTO.ToUserResponse(user, h.storageUC))
}

// AddCrush godoc
// @Summary Simulate adding a crush package to user
// @Tags admin
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body AdminConsumableRequest true "Crush Request"
// @Success 200 {object} userDTO.UserResponse
// @Router /admin/consumables/crush [post]
func (h *AdminHandler) AddCrush(c *gin.Context) {
	var req AdminConsumableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	uid, err := uuid.Parse(req.UserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user_id", err.Error())
		return
	}

	pkgID, err := uuid.Parse(req.PackageID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid package_id", err.Error())
		return
	}

	user, err := h.adminSvc.AddConsumable(c.Request.Context(), uid, pkgID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to add crush", err.Error())
		return
	}

	response.OK(c, userDTO.ToUserResponse(user, h.storageUC))
}

// GetAllConfigs godoc
// @Summary Get all configs
// @Tags admin
// @Security Bearer
// @Success 200 {object} response.BaseResponse
// @Router /admin/configs [get]
func (h *AdminHandler) GetAllConfigs(c *gin.Context) {
	configs := h.configSvc.GetAllCached(c.Request.Context())
	response.OK(c, configs)
}

// ReloadConfigs godoc
// @Summary Reload settings from database to RAM
// @Tags admin
// @Security Bearer
// @Success 200 {object} map[string]string
// @Router /admin/configs/reload [post]
func (h *AdminHandler) ReloadConfigs(c *gin.Context) {
	if err := h.configSvc.LoadConfigs(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload configs: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Configs reloaded successfully"})
}

// ResetConfigs godoc
// @Summary WIPE and Reset all configs to code defaults
// @Tags admin
// @Security Bearer
// @Success 200 {object} map[string]string
// @Router /admin/configs/reset [post]
func (h *AdminHandler) ResetConfigs(c *gin.Context) {
	// 1. Reset DB
	if err := h.configSvc.ResetDB(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset configs DB: " + err.Error()})
		return
	}

	// 2. Re-seed (This will re-populate app_configs, subscription_plans, etc. but since we only deleted app_configs it is safe)
	if err := seeds.SeedMasterData(h.db); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to re-seed master data: " + err.Error()})
		return
	}

	// 3. Reload RAM
	if err := h.configSvc.LoadConfigs(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload configs: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Configs reset to defaults and reloaded successfully"})
}
