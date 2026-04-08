package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	base "github.com/pipigendut/dating-backend/internal/delivery/http/dto"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/repository"
)

type DeviceHandler struct {
	repo      repository.DeviceRepository
	notifRepo repository.NotificationRepository
}

func NewDeviceHandler(repo repository.DeviceRepository, notifRepo repository.NotificationRepository) *DeviceHandler {
	return &DeviceHandler{
		repo:      repo,
		notifRepo: notifRepo,
	}
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
		base.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
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

	// 1. Upsert device
	if err := h.repo.UpsertDevice(c.Request.Context(), device); err != nil {
		base.Error(c, http.StatusInternalServerError, "Failed to register device", err.Error())
		return
	}

	// 2. Activate all notification settings for the user (as requested by user)
	if err := h.notifRepo.ActivateAllUserSettings(c.Request.Context(), userID); err != nil {
		base.Error(c, http.StatusInternalServerError, "Failed to activate all user settings", err.Error())
		return
	}

	base.OK(c, gin.H{"message": "Device registered successfully"})
}

type DeactivateDeviceRequest struct {
	DeviceID string `json:"device_id" binding:"required"`
}

func (h *DeviceHandler) DeactivateDevice(c *gin.Context) {
	var req DeactivateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	userID := c.MustGet("userID").(uuid.UUID)

	// 1. Deactivate device
	if err := h.repo.DeactivateDevice(c.Request.Context(), userID, req.DeviceID); err != nil {
		base.Error(c, http.StatusInternalServerError, "Failed to deactivate device", err.Error())
		return
	}

	// 2. Disable all notification settings for the user (as requested by user)
	if err := h.notifRepo.DeactivateAllUserSettings(c.Request.Context(), userID); err != nil {
		base.Error(c, http.StatusInternalServerError, "Failed to deactivate all user settings", err.Error())
		return
	}

	base.OK(c, gin.H{"message": "Device deactivated and all notifications disabled"})
}

type UpdateFCMTokenRequest struct {
	DeviceID string `json:"device_id" binding:"required"`
	FCMToken string `json:"fcm_token" binding:"required"`
}

func (h *DeviceHandler) UpdateFCMToken(c *gin.Context) {
	var req UpdateFCMTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	userID := c.MustGet("userID").(uuid.UUID)

	if err := h.repo.UpdateFCMToken(c.Request.Context(), userID, req.DeviceID, req.FCMToken); err != nil {
		base.Error(c, http.StatusInternalServerError, "Failed to update FCM token", err.Error())
		return
	}

	base.OK(c, gin.H{"message": "FCM token updated successfully"})
}
