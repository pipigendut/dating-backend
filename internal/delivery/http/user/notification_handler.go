package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/delivery/http/response"
	"github.com/pipigendut/dating-backend/internal/services"
)

type NotificationHandler struct {
	notifConfigService services.NotificationConfigService
}

func NewNotificationHandler(r *gin.RouterGroup, notifConfigService services.NotificationConfigService, authMiddleware gin.HandlerFunc) {
	handler := &NotificationHandler{
		notifConfigService: notifConfigService,
	}
	
	notif := r.Group("/users/notifications")
	notif.Use(authMiddleware)
	{
		notif.GET("", handler.GetUserSettings)
		notif.POST("", handler.UpdateUserSetting)
	}
}

type NotificationSettingRequest struct {
	NotificationSettingID uuid.UUID `json:"notification_setting_id" binding:"required"`
	IsEnable              bool      `json:"is_enable"`
}

type NotificationSettingResponse struct {
	ID           uuid.UUID `json:"id"`
	Type         string    `json:"type"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	IsEnable     bool      `json:"is_enable"`
	IsUserEnable bool      `json:"is_user_enable"`
}

// GetUserSettings godoc
// @Summary      Get user notification settings
// @Description  Fetches the list of all master notification settings with the current user's preferences.
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.BaseResponse{data=[]NotificationSettingResponse} "List of notification settings"
// @Failure      401  {object}  response.BaseResponse "Unauthorized"
// @Failure      500  {object}  response.BaseResponse "Internal server error"
// @Router       /users/notifications [get]
func (h *NotificationHandler) GetUserSettings(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	
	settings, err := h.notifConfigService.GetUserSettings(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch notification settings", err.Error())
		return
	}
	
	var res []NotificationSettingResponse
	for _, s := range settings {
		res = append(res, NotificationSettingResponse{
			ID:           s.NotificationSettingID,
			Type:         s.NotificationSetting.Type,
			Title:        s.NotificationSetting.Title,
			Description:  s.NotificationSetting.Description,
			IsEnable:     s.NotificationSetting.IsEnable,
			IsUserEnable: s.IsEnable,
		})
	}
	
	response.OK(c, res)
}

// UpdateUserSetting godoc
// @Summary      Update a user notification setting
// @Description  Enables or disables a specific notification type for the authenticated user using upsert logic.
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request     body      NotificationSettingRequest true "Update payload"
// @Success      200  {object}  response.BaseResponse "Success"
// @Failure      400  {object}  response.BaseResponse "Invalid request"
// @Failure      401  {object}  response.BaseResponse "Unauthorized"
// @Failure      500  {object}  response.BaseResponse "Internal server error"
// @Router       /users/notifications [post]
func (h *NotificationHandler) UpdateUserSetting(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	
	var req NotificationSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}
	
	if err := h.notifConfigService.UpdateUserSetting(c.Request.Context(), userID, req.NotificationSettingID, req.IsEnable); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update notification setting", err.Error())
		return
	}
	
	response.OK(c, nil)
}

