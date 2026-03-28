package user

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/delivery/http/response"
	"github.com/pipigendut/dating-backend/internal/services"
)

type InviteHandler struct {
	groupSvc       services.GroupService
	storageService *services.StorageService
}

func NewInviteHandler(r *gin.RouterGroup, groupSvc services.GroupService, storageService *services.StorageService, authMiddleware gin.HandlerFunc) {
	handler := &InviteHandler{
		groupSvc:       groupSvc,
		storageService: storageService,
	}

	// Well-known endpoints (Must be at the root of the domain, usually)
	// But our router might be nested. In main.go we should register these at the root.
	
	groups := r.Group("/groups")
	groups.Use(authMiddleware)
	{
		groups.POST("/:id/invite-link", handler.GenerateInviteLink)
	}

	invites := r.Group("/group-invites")
	{
		invites.GET("/validate", handler.ValidateInvite)
		
		protected := invites.Group("")
		protected.Use(authMiddleware)
		protected.POST("/accept", handler.AcceptInvite)
	}
}

// GenerateInviteLink godoc
// @Summary      Generate group invite link
// @Description  Creates a secure, single-use invite link for a group.
// @Tags         groups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Group ID"
// @Success      200  {object}  response.BaseResponse{data=string} "Invite Link"
// @Router       /groups/{id}/invite-link [post]
func (h *InviteHandler) GenerateInviteLink(c *gin.Context) {
	groupIDStr := c.Param("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid Group ID", nil)
		return
	}

	userID := c.MustGet("userID").(uuid.UUID)

	// Check if user is member
	isMember, err := h.groupSvc.IsGroupMember(c.Request.Context(), groupID, userID)
	if err != nil || !isMember {
		response.Error(c, http.StatusForbidden, "Only group members can generate invite links", nil)
		return
	}

	invite, err := h.groupSvc.GenerateInviteLink(c.Request.Context(), groupID, userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to generate invite", err.Error())
		return
	}

	// Construct full link using NGROK_URL from env
	ngrokURL := os.Getenv("NGROK_URL")
	if ngrokURL == "" {
		ngrokURL = "http://localhost:8080"
	}
	
	inviteLink := ngrokURL + "/invite?token=" + invite.Token

	response.OK(c, inviteLink)
}

// ValidateInvite godoc
// @Summary      Validate invite token
// @Description  Checks if an invite token is valid, not expired, and not used.
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        token  query     string  true  "Invite Token"
// @Success      200  {object}  response.BaseResponse{data=map[string]interface{}}
// @Router       /group-invites/validate [get]
func (h *InviteHandler) ValidateInvite(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		response.Error(c, http.StatusBadRequest, "Token is required", nil)
		return
	}

	invite, err := h.groupSvc.ValidateInvite(c.Request.Context(), token)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	response.OK(c, gin.H{
		"group_id":     invite.GroupID,
		"group_name":   invite.Group.Name,
		"inviter_name": invite.Inviter.FullName,
		"is_valid":     true,
	})
}

// AcceptInvite godoc
// @Summary      Accept group invitation
// @Description  Uses a valid token to join a group and update the user's active entity.
// @Tags         groups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body map[string]string true "Token payload"
// @Success      200  {object}  response.BaseResponse
// @Router       /group-invites/accept [post]
func (h *InviteHandler) AcceptInvite(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	userID := c.MustGet("userID").(uuid.UUID)

	err := h.groupSvc.AcceptInvite(c.Request.Context(), req.Token, userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	response.OK(c, "Successfully joined the group")
}
