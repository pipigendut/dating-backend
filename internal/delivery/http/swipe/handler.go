package swipe

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/delivery/http/response"
	"github.com/pipigendut/dating-backend/internal/delivery/http/user"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/services"
)

func NewSwipeHandler(r *gin.RouterGroup, swipeSvc services.SwipeService, storageUC storageUsecase, authMiddleware gin.HandlerFunc) {
	handler := &SwipeHandler{
		swipeService: swipeSvc,
		storageUC:    storageUC,
	}

	swipeGroup := r.Group("/swipe")
	swipeGroup.Use(authMiddleware)
	{
		swipeGroup.GET("/candidates", handler.GetCandidates)
		swipeGroup.POST("/", handler.Swipe)
		swipeGroup.GET("/likes", handler.GetIncomingLikes)
		swipeGroup.POST("/undo", handler.UndoSwipe)
	}
}

// storageUsecase is a minimal interface for photo URL resolution (avoids circular imports)
type storageUsecase interface {
	GetPublicURL(key string) string
}

type SwipeHandler struct {
	swipeService services.SwipeService
	storageUC    storageUsecase
}

// resolvePhotoURLs converts raw S3 file keys in Photos to full public URLs
func (h *SwipeHandler) resolvePhotoURLs(u *entities.User) {
	if h.storageUC == nil {
		return
	}
	for i := range u.Photos {
		if u.Photos[i].URL != "" {
			u.Photos[i].URL = h.storageUC.GetPublicURL(u.Photos[i].URL)
		}
	}
}

// GetCandidates godoc
// @Summary      Get list of users for swipe discovery
// @Description  Fetches a weighted-random list of active users that the current user hasn't swiped on yet, applying cooldowns and priority scoring.
// @Tags         swipe
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.BaseResponse{data=[]user.UserResponse} "List of swipe candidates"
// @Failure      500  {object}  response.BaseResponse "Internal server error"
// @Router       /swipe/candidates [get]
func (h *SwipeHandler) GetCandidates(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	candidates, err := h.swipeService.GetSwipeCandidates(c.Request.Context(), userID, 10)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get swipe candidates", err.Error())
		return
	}

	// Record impressions for these candidates immediately since we're returning them to the client
	// In a real high-scale app, the client might report back which ones were actually viewed
	var shownIDs []uuid.UUID
	var respCandidates []user.UserResponse
	for _, u := range candidates {
		shownIDs = append(shownIDs, u.ID)
		uCopy := u // capture loop variable cleanly
		h.resolvePhotoURLs(&uCopy)
		respCandidates = append(respCandidates, user.ToUserResponse(&uCopy))
	}

	// Fire and forget impression recording
	go func() {
		// Create a background context since the request context will be cancelled when response returns
		bgCtx := context.Background()
		h.swipeService.RecordImpressions(bgCtx, userID, shownIDs)
	}()

	response.OK(c, respCandidates)
}

// Swipe godoc
// @Summary      Record a swipe action (LIKE, CRUSH or DISLIKE)
// @Description  Records a user's swipe interaction with another user. If it's a mutual LIKE or CRUSH, it returns a Match ID.
// @Tags         swipe
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body SwipeRequest true "Swipe action details"
// @Success      200  {object}  response.BaseResponse{data=MatchResponse} "Swipe recorded successfully, returns match status"
// @Failure      400  {object}  response.BaseResponse "Invalid request"
// @Failure      500  {object}  response.BaseResponse "Internal server error"
// @Router       /swipe [post]
func (h *SwipeHandler) Swipe(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var req SwipeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	match, matchedUser, err := h.swipeService.CreateSwipe(c.Request.Context(), userID, req.SwipedID, req.Direction)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to record swipe", err.Error())
		return
	}

	if match != nil {
		var matchedUserResp *user.UserResponse
		if matchedUser != nil {
			h.resolvePhotoURLs(matchedUser)
			ur := user.ToUserResponse(matchedUser)
			matchedUserResp = &ur
		}
		response.OK(c, MatchResponse{
			IsMatch:     true,
			MatchID:     match.ID,
			MatchedUser: matchedUserResp,
		})
		return
	}

	response.OK(c, MatchResponse{IsMatch: false})
}

// GetIncomingLikes godoc
// @Summary      Get list of incoming likes and crushes
// @Description  Fetches a list of users who have liked or crushed on the current user, ordered by priority score (Crushes first, then Premium likes, then standard likes).
// @Tags         swipe
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.BaseResponse{data=[]IncomingLikeResponse} "List of incoming likes"
// @Failure      500  {object}  response.BaseResponse "Internal server error"
// @Router       /swipe/likes [get]
func (h *SwipeHandler) GetIncomingLikes(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	likes, err := h.swipeService.GetIncomingLikes(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get incoming likes", err.Error())
		return
	}

	var respLikes []IncomingLikeResponse
	for _, like := range likes {
		userCopy := like.User
		h.resolvePhotoURLs(&userCopy)
		userResp := user.ToUserResponse(&userCopy)

		respLikes = append(respLikes, IncomingLikeResponse{
			User:          userResp,
			IsCrush:       like.IsCrush,
			PriorityScore: like.PriorityScore,
			SwipeTime:     like.CreatedAt.Format(time.RFC3339),
		})
	}

	response.OK(c, respLikes)
}

// UndoSwipe godoc
// @Summary      Undo the last swipe action
// @Description  Reverts the most recent swipe (like, dislike, or crush). If it was a match, the match is also removed. Returns the details of the undone user so they can be shown again in the UI.
// @Tags         swipe
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.BaseResponse{data=user.UserResponse} "Successfully reverted the swipe"
// @Failure      400  {object}  response.BaseResponse "No swipe history or daily limit reached"
// @Failure      500  {object}  response.BaseResponse "Internal server error"
// @Router       /swipe/undo [post]
func (h *SwipeHandler) UndoSwipe(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	undoneUser, err := h.swipeService.UndoLastSwipe(c.Request.Context(), userID)
	if err != nil {
		if err.Error() == "no swipe history found to undo" || err.Error()[:17] == "daily undo limit" {
			response.Error(c, http.StatusBadRequest, "Cannot undo swipe", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to undo swipe", err.Error())
		return
	}

	h.resolvePhotoURLs(undoneUser)
	userResp := user.ToUserResponse(undoneUser)

	response.OK(c, userResp)
}
