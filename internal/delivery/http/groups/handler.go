package groups

import (
	"errors"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/delivery/http/response"
	"github.com/pipigendut/dating-backend/internal/services"
	"gorm.io/gorm"
)

type GroupHandler struct {
	groupService   services.GroupService
	storageService *services.StorageService
}

func NewGroupHandler(r *gin.RouterGroup, groupSvc services.GroupService, storageService *services.StorageService, authMiddleware gin.HandlerFunc) *GroupHandler {
	handler := &GroupHandler{
		groupService:   groupSvc,
		storageService: storageService,
	}

	// Group Management (Authenticated)
	groups := r.Group("/groups")
	groups.Use(authMiddleware)
	{
		groups.POST("", handler.CreateGroup)
		groups.GET("/me", handler.GetMyGroup)
		groups.DELETE("/:id", handler.DisbandGroup)
		groups.POST("/:id/leave", handler.LeaveGroup)
		groups.DELETE("/:id/members/:user_id", handler.KickMember)
		groups.POST("/:id/invite", handler.InviteToGroup)
		groups.POST("/:id/invite-link", handler.GenerateInviteLink)
	}

	// Group Invitations (Mixed Auth)
	invites := r.Group("/group-invites")
	{
		invites.GET("/validate", handler.ValidateInvite)
		
		protected := invites.Group("")
		protected.Use(authMiddleware)
		protected.POST("/accept", handler.AcceptInvite)
	}

	return handler
}

// CreateGroup godoc
// @Summary      Create a group
// @Description  Creates a new group entity with the authenticated user as the owner.
// @Tags         groups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body CreateGroupRequest true "Group details"
// @Success      200  {object}  response.BaseResponse{data=entities.Group} "Created group details"
// @Router       /groups [post]
func (h *GroupHandler) CreateGroup(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	group, err := h.groupService.CreateGroup(c.Request.Context(), userID, req.Name)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create group", err.Error())
		return
	}

	response.OK(c, group)
}

// GetMyGroup godoc
// @Summary      Get user's group
// @Description  Fetches the details of the single group the authenticated user belongs to, including its members.
// @Tags         groups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.BaseResponse{data=response.GroupResponse} "Group details"
// @Failure      404  {object}  response.BaseResponse "Group not found"
// @Router       /groups/me [get]
func (h *GroupHandler) GetMyGroup(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	group, err := h.groupService.GetMyGroup(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.OK(c, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to fetch group", err.Error())
		return
	}

	// Map members
	members := make([]response.UserResponse, len(group.Members))
	for i, m := range group.Members {
		if m.User != nil {
			members[i] = response.ToUserResponse(m.User, h.storageService)
		}
	}

	resp := response.GroupResponse{
		ID:        group.ID,
		EntityID:  group.EntityID,
		Name:      group.Name,
		CreatedBy: group.CreatedBy,
		Members:   members,
	}

	response.OK(c, resp)
}

// InviteToGroup godoc
// @Summary      Invite user to group
// @Description  Invites another user to join a specific group.
// @Tags         groups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Group ID"
// @Param        request body InviteToGroupRequest true "Invite details"
// @Success      200  {object}  response.BaseResponse "Successfully invited"
// @Router       /groups/{id}/invite [post]
func (h *GroupHandler) InviteToGroup(c *gin.Context) {
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid group ID", err.Error())
		return
	}

	var req InviteToGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.groupService.InviteToGroup(c.Request.Context(), groupID, req.UserID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to invite to group", err.Error())
		return
	}

	response.OK(c, nil)
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
func (h *GroupHandler) GenerateInviteLink(c *gin.Context) {
	groupIDStr := c.Param("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid Group ID", nil)
		return
	}

	userID := c.MustGet("userID").(uuid.UUID)

	// Check if user is member
	isMember, err := h.groupService.IsGroupMember(c.Request.Context(), groupID, userID)
	if err != nil || !isMember {
		response.Error(c, http.StatusForbidden, "Only group members can generate invite links", nil)
		return
	}

	invite, err := h.groupService.GenerateInviteLink(c.Request.Context(), groupID, userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to generate invite", err.Error())
		return
	}

	// Always generate HTTPS Universal Links for sharing
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	inviteLink := baseURL + "/invite/" + invite.Token
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
func (h *GroupHandler) ValidateInvite(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		response.Error(c, http.StatusBadRequest, "Token is required", nil)
		return
	}

	invite, err := h.groupService.ValidateInvite(c.Request.Context(), token)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	response.OK(c, gin.H{
		"group_id":   invite.GroupID,
		"group_name": invite.Group.Name,
		"is_valid":   true,
	})
}

// AcceptInvite godoc
// @Summary      Accept group invitation
// @Description  Uses a valid token to join a group and update the user's active entity.
// @Tags         groups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body AcceptInviteRequest true "Token payload"
// @Success      200  {object}  response.BaseResponse
// @Router       /group-invites/accept [post]
func (h *GroupHandler) AcceptInvite(c *gin.Context) {
	var req AcceptInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	userID := c.MustGet("userID").(uuid.UUID)

	err := h.groupService.AcceptInvite(c.Request.Context(), req.Token, userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	response.OK(c, "Successfully joined the group")
}

// HandleInviteRedirect handles the root /invite (and /invite/:token) GET requests.
// It serves a beautiful landing page that tries to open the app via custom scheme.
func (h *GroupHandler) HandleInviteRedirect(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		token = c.Query("token")
	}

	// Fetch invite details to show on the landing page if possible
	invite, _ := h.groupService.ValidateInvite(c.Request.Context(), token)
	groupName := "a Group"
	inviterName := "a Friend"
	if invite != nil {
		groupName = invite.Group.Name
		inviterName = invite.Inviter.FullName
	}

	customScheme := "swipee://invite/" + token
	playStoreURL := "https://play.google.com/store/apps/details?id=com.swipee"
	appStoreURL := "https://apps.apple.com/app/idYOUR_APP_ID"

	html := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Join ` + groupName + ` on Swipee</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
            background: linear-gradient(135deg, #df2c2c 0%, #ff5e62 100%);
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            height: 100vh;
            margin: 0;
            color: white;
            text-align: center;
            padding: 20px;
        }
        .container {
            background: rgba(255, 255, 255, 0.1);
            backdrop-filter: blur(10px);
            padding: 40px;
            border-radius: 24px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.2);
            max-width: 400px;
            width: 100%;
        }
        .logo {
            font-size: 48px;
            font-weight: 800;
            margin-bottom: 24px;
            letter-spacing: -2px;
        }
        h1 { font-size: 24px; margin-bottom: 8px; }
        p { font-size: 16px; opacity: 0.9; margin-bottom: 32px; }
        .btn {
            background: white;
            color: #df2c2c;
            padding: 16px 32px;
            border-radius: 100px;
            text-decoration: none;
            font-weight: bold;
            font-size: 18px;
            display: block;
            transition: transform 0.2s;
            box-shadow: 0 4px 15px rgba(0,0,0,0.1);
        }
        .btn:active { transform: scale(0.95); }
        .footer { margin-top: 32px; font-size: 14px; opacity: 0.7; }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">Swipee</div>
        <h1>You're Invited!</h1>
        <p><strong>` + inviterName + `</strong> wants you to join the group <strong>` + groupName + `</strong>.</p>
        <a href="` + customScheme + `" class="btn" id="openBtn">Join Group in App</a>
    </div>
    <div class="footer">
        Don't have the app? <br>
        <a href="` + playStoreURL + `" style="color: white;">Download for Android</a> or 
        <a href="` + appStoreURL + `" style="color: white;">iOS</a>
    </div>
    <script>
        // Optional: Auto-trigger after a delay
        // window.location.href = "` + customScheme + `";
    </script>
</body>
</html>
`
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// KickMember godoc
// @Summary      Kick group member
// @Description  Allows owner to kick a specific member
// @Tags         groups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        group_id path string true "Group ID"
// @Param        user_id path string true "User ID to Kick"
// @Success      200  {object}  response.BaseResponse
// @Router       /groups/{group_id}/members/{user_id} [delete]
func (h *GroupHandler) KickMember(c *gin.Context) {
	ownerID := c.MustGet("userID").(uuid.UUID)
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid group ID", err.Error())
		return
	}
	targetUserID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	if err := h.groupService.KickMember(c.Request.Context(), groupID, ownerID, targetUserID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to kick member", err.Error())
		return
	}

	response.OK(c, nil)
}

// LeaveGroup godoc
// @Summary      Leave group
// @Description  Member leaves their group
// @Tags         groups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        group_id path string true "Group ID"
// @Success      200  {object}  response.BaseResponse
// @Router       /groups/{group_id}/leave [post]
func (h *GroupHandler) LeaveGroup(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid group ID", err.Error())
		return
	}

	if err := h.groupService.LeaveGroup(c.Request.Context(), groupID, userID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to leave group", err.Error())
		return
	}

	response.OK(c, nil)
}

// DisbandGroup godoc
// @Summary      Disband group
// @Description  Owner deletes the entire group
// @Tags         groups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        group_id path string true "Group ID"
// @Success      200  {object}  response.BaseResponse
// @Router       /groups/{group_id} [delete]
func (h *GroupHandler) DisbandGroup(c *gin.Context) {
	ownerID := c.MustGet("userID").(uuid.UUID)
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid group ID", err.Error())
		return
	}

	if err := h.groupService.DisbandGroup(c.Request.Context(), groupID, ownerID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to disband group", err.Error())
		return
	}

	response.OK(c, nil)
}

