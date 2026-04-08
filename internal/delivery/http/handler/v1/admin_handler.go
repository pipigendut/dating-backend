package v1

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	base "github.com/pipigendut/dating-backend/internal/delivery/http/dto"
	dtov1 "github.com/pipigendut/dating-backend/internal/delivery/http/dto/v1"
	"github.com/pipigendut/dating-backend/internal/delivery/http/mapper"
	"github.com/pipigendut/dating-backend/internal/infra/seeds"
	"github.com/pipigendut/dating-backend/internal/repository"
	"github.com/pipigendut/dating-backend/internal/services"
	"gorm.io/gorm"
)

type AdminHandler struct {
	db             *gorm.DB
	configSvc      services.ConfigService
	adminSvc       services.AdminService
	userRepo       repository.UserRepository
	storageService *services.StorageService
}

var _ = dtov1.UserResponse{}

func NewAdminHandler(db *gorm.DB, configSvc services.ConfigService, adminSvc services.AdminService, userRepo repository.UserRepository, storageService *services.StorageService) *AdminHandler {
	return &AdminHandler{
		db:             db,
		configSvc:      configSvc,
		adminSvc:       adminSvc,
		userRepo:       userRepo,
		storageService: storageService,
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
			base.Error(c, http.StatusUnauthorized, "User not identified", "")
			c.Abort()
			return
		}

		userID := val.(uuid.UUID)
		user, err := h.userRepo.GetByID(userID)
		if err != nil {
			base.Error(c, http.StatusInternalServerError, "Failed to verify whitelist", err.Error())
			c.Abort()
			return
		}

		var allowedEmails []string
		if strings.HasPrefix(whitelistStr, "[") {
			if err := json.Unmarshal([]byte(whitelistStr), &allowedEmails); err != nil {
				base.Error(c, http.StatusInternalServerError, "Failed to parse whitelist", err.Error())
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
			base.Error(c, http.StatusForbidden, "Email not whitelisted for admin actions", userEmail)
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
// @Success 200 {object} base.BaseResponse{data=dtov1.UserResponse}
// @Router /admin/subscribe [post]
func (h *AdminHandler) SubscribeUser(c *gin.Context) {
	var req AdminSubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	uid, err := uuid.Parse(req.UserID)
	if err != nil {
		base.Error(c, http.StatusBadRequest, "Invalid user_id", err.Error())
		return
	}

	pid, err := uuid.Parse(req.PlanID)
	if err != nil {
		base.Error(c, http.StatusBadRequest, "Invalid plan_id", err.Error())
		return
	}

	user, err := h.adminSvc.SubscribeUser(c.Request.Context(), uid, pid)
	if err != nil {
		base.Error(c, http.StatusInternalServerError, "Failed to subscribe user", err.Error())
		return
	}

	base.OK(c, mapper.ToUserResponse(user, h.storageService))
}

// AddBoost godoc
// @Summary Simulate adding a boost package to user
// @Tags admin
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body AdminConsumableRequest true "Boost Request"
// @Success 200 {object} base.BaseResponse{data=dtov1.UserResponse}
// @Router /admin/consumables/boost [post]
func (h *AdminHandler) AddBoost(c *gin.Context) {
	var req AdminConsumableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	uid, err := uuid.Parse(req.UserID)
	if err != nil {
		base.Error(c, http.StatusBadRequest, "Invalid user_id", err.Error())
		return
	}

	pkgID, err := uuid.Parse(req.PackageID)
	if err != nil {
		base.Error(c, http.StatusBadRequest, "Invalid package_id", err.Error())
		return
	}

	// Note: ItemType check could be added if needed, but AdminService.AddConsumable handles any package.
	user, err := h.adminSvc.AddConsumable(c.Request.Context(), uid, pkgID)
	if err != nil {
		base.Error(c, http.StatusInternalServerError, "Failed to add boost", err.Error())
		return
	}

	base.OK(c, mapper.ToUserResponse(user, h.storageService))
}

// AddCrush godoc
// @Summary Simulate adding a crush package to user
// @Tags admin
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body AdminConsumableRequest true "Crush Request"
// @Success 200 {object} base.BaseResponse{data=dtov1.UserResponse}
// @Router /admin/consumables/crush [post]
func (h *AdminHandler) AddCrush(c *gin.Context) {
	var req AdminConsumableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	uid, err := uuid.Parse(req.UserID)
	if err != nil {
		base.Error(c, http.StatusBadRequest, "Invalid user_id", err.Error())
		return
	}

	pkgID, err := uuid.Parse(req.PackageID)
	if err != nil {
		base.Error(c, http.StatusBadRequest, "Invalid package_id", err.Error())
		return
	}

	user, err := h.adminSvc.AddConsumable(c.Request.Context(), uid, pkgID)
	if err != nil {
		base.Error(c, http.StatusInternalServerError, "Failed to add crush", err.Error())
		return
	}

	base.OK(c, mapper.ToUserResponse(user, h.storageService))
}

// GetAllConfigs godoc
// @Summary Get all configs
// @Tags admin
// @Security Bearer
// @Success 200 {object} base.BaseResponse
// @Router /admin/configs [get]
func (h *AdminHandler) GetAllConfigs(c *gin.Context) {
	configs := h.configSvc.GetAllCached(c.Request.Context())
	base.OK(c, configs)
}

// ReloadConfigs godoc
// @Summary Reload settings from database to RAM
// @Tags admin
// @Security Bearer
// @Success 200 {object} map[string]string
// @Router /admin/configs/reload [post]
func (h *AdminHandler) ReloadConfigs(c *gin.Context) {
	if err := h.configSvc.LoadConfigs(c.Request.Context()); err != nil {
		base.Error(c, http.StatusInternalServerError, "Failed to reload configs", err.Error())
		return
	}
	base.OK(c, "Configs reloaded successfully")
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
		base.Error(c, http.StatusInternalServerError, "Failed to reset configs DB", err.Error())
		return
	}

	// 2. Re-seed (This will re-populate app_configs, subscription_plans, etc. but since we only deleted app_configs it is safe)
	if err := seeds.SeedMasterData(h.db); err != nil {
		base.Error(c, http.StatusInternalServerError, "Failed to re-seed master data", err.Error())
		return
	}

	// 3. Reload RAM
	if err := h.configSvc.LoadConfigs(c.Request.Context()); err != nil {
		base.Error(c, http.StatusInternalServerError, "Failed to reload configs", err.Error())
		return
	}

	base.OK(c, "Configs reset to defaults and reloaded successfully")
}
