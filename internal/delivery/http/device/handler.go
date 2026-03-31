package device

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/delivery/http/response"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/repository"
)

type DeviceHandler struct {
	repo      repository.DeviceRepository
	notifRepo repository.NotificationRepository
}

func NewDeviceHandler(r *gin.RouterGroup, repo repository.DeviceRepository, notifRepo repository.NotificationRepository, authMiddleware gin.HandlerFunc) *DeviceHandler {
	h := &DeviceHandler{
		repo:      repo,
		notifRepo: notifRepo,
	}
	
	devices := r.Group("/devices")
	devices.Use(authMiddleware)
	{
		devices.POST("/register", h.RegisterDevice)
		devices.PATCH("/fcm-token", h.UpdateFCMToken)
		devices.POST("/deactivate", h.DeactivateDevice)
	}

	
	return h
}

type RegisterDeviceRequest struct {
	DeviceID    string `json:"device_id" binding:"required"`
	DeviceName  string `json:"device_name"`
	DeviceModel string `json:"device_model"`
	OSVersion   string `json:"os_version"`
	AppVersion  string `json:"app_version"`
	FCMToken    string `json:"fcm_token"`
}

func (h *DeviceHandler) RegisterDevice(c *gin.Context) {
	var req RegisterDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	userID := c.MustGet("userID").(uuid.UUID)
	
	device := &entities.Device{
		UserID:      userID,
		DeviceID:    req.DeviceID,
		DeviceName:  req.DeviceName,
		DeviceModel: req.DeviceModel,
		OSVersion:   req.OSVersion,
		AppVersion:  req.AppVersion,
		FCMToken:    &req.FCMToken,
		IsActive:    true,
		LastIP:      c.ClientIP(),
	}

	if err := h.repo.UpsertDevice(c.Request.Context(), device); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to register device", err.Error())
		return
	}

	response.OK(c, gin.H{"message": "Device registered successfully"})
}

type UpdateFCMTokenRequest struct {
	DeviceID string `json:"device_id" binding:"required"`
	FCMToken string `json:"fcm_token" binding:"required"`
}

func (h *DeviceHandler) UpdateFCMToken(c *gin.Context) {
	var req UpdateFCMTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	userID := c.MustGet("userID").(uuid.UUID)

	if err := h.repo.UpdateFCMToken(c.Request.Context(), userID, req.DeviceID, req.FCMToken); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update FCM token", err.Error())
		return
	}

	response.OK(c, gin.H{"message": "FCM token updated successfully"})
}

type DeactivateDeviceRequest struct {
	DeviceID string `json:"device_id" binding:"required"`
}

func (h *DeviceHandler) DeactivateDevice(c *gin.Context) {
	var req DeactivateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	userID := c.MustGet("userID").(uuid.UUID)

	// 1. Deactivate device
	if err := h.repo.DeactivateDevice(c.Request.Context(), userID, req.DeviceID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to deactivate device", err.Error())
		return
	}

	// 2. Disable all notification settings for the user (as requested by user)
	if err := h.notifRepo.DeactivateAllUserSettings(c.Request.Context(), userID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to deactivate all user settings", err.Error())
		return
	}

	response.OK(c, gin.H{"message": "Device deactivated and all notifications disabled"})
}

