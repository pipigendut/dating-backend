package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pipigendut/dating-backend/internal/infra/seeds"
	"github.com/pipigendut/dating-backend/internal/services"
	"gorm.io/gorm"
)

type AdminHandler struct {
	db        *gorm.DB
	configSvc services.ConfigService
}

func NewAdminHandler(db *gorm.DB, configSvc services.ConfigService) *AdminHandler {
	return &AdminHandler{
		db:        db,
		configSvc: configSvc,
	}
}

func (h *AdminHandler) RegisterRoutes(rg *gin.RouterGroup) {
	admin := rg.Group("/admin")
	{
		configs := admin.Group("/configs")
		{
			configs.POST("/reload", h.ReloadConfigs)
			configs.POST("/reset", h.ResetConfigs)
		}
	}
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
