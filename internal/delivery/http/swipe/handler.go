package swipe

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/delivery/http/response"
	"github.com/pipigendut/dating-backend/internal/services"
)

func NewSwipeHandler(r *gin.RouterGroup, swipeSvc services.SwipeService, authMiddleware gin.HandlerFunc) {
	handler := &SwipeHandler{
		swipeService: swipeSvc,
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

type SwipeHandler struct {
	swipeService services.SwipeService
}

// GetCandidates godoc
// @Summary      Get list of users for swipe discovery
// @Description  Fetches a weighted-random list of active users that the current user hasn't swiped on yet, applying cooldowns and priority scoring.
// @Tags         swipe
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.BaseResponse{data=[]UserSwipeProfileResponse} "List of swipe candidates"
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
	var respCandidates []UserSwipeProfileResponse
	for _, user := range candidates {
		shownIDs = append(shownIDs, user.ID)

		// Map photos
		var photos []PhotoDTO
		for _, p := range user.Photos {
			photos = append(photos, PhotoDTO{
				ID:        p.ID,
				URL:       p.URL,
				IsMain:    p.IsMain,
				SortOrder: p.SortOrder,
			})
		}

		respCandidates = append(respCandidates, UserSwipeProfileResponse{
			ID:              user.ID,
			FullName:        user.FullName,
			Bio:             user.Bio,
			HeightCM:        user.HeightCM,
			LocationCity:    user.LocationCity,
			LocationCountry: user.LocationCountry,
			Photos:          photos,
		})
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

	match, err := h.swipeService.CreateSwipe(c.Request.Context(), userID, req.SwipedID, req.Direction)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to record swipe", err.Error())
		return
	}

	if match != nil {
		response.OK(c, MatchResponse{IsMatch: true, MatchID: match.ID})
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
		var photos []PhotoDTO
		for _, p := range like.User.Photos {
			photos = append(photos, PhotoDTO{
				ID:        p.ID,
				URL:       p.URL,
				IsMain:    p.IsMain,
				SortOrder: p.SortOrder,
			})
		}

		userResp := UserSwipeProfileResponse{
			ID:              like.User.ID,
			FullName:        like.User.FullName,
			Bio:             like.User.Bio,
			HeightCM:        like.User.HeightCM,
			LocationCity:    like.User.LocationCity,
			LocationCountry: like.User.LocationCountry,
			Photos:          photos,
		}

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
// @Success      200  {object}  response.BaseResponse{data=UserSwipeProfileResponse} "Successfully reverted the swipe"
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

	// Map returned user to response DTO
	var photos []PhotoDTO
	for _, p := range undoneUser.Photos {
		photos = append(photos, PhotoDTO{
			ID:        p.ID,
			URL:       p.URL,
			IsMain:    p.IsMain,
			SortOrder: p.SortOrder,
		})
	}

	userResp := UserSwipeProfileResponse{
		ID:              undoneUser.ID,
		FullName:        undoneUser.FullName,
		Bio:             undoneUser.Bio,
		HeightCM:        undoneUser.HeightCM,
		LocationCity:    undoneUser.LocationCity,
		LocationCountry: undoneUser.LocationCountry,
		Photos:          photos,
	}

	response.OK(c, userResp)
}
