package groups

import "github.com/google/uuid"

type CreateGroupRequest struct {
	Name string `json:"name" binding:"required"`
}

type InviteToGroupRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
}

type AcceptInviteRequest struct {
	Token string `json:"token" binding:"required"`
}
