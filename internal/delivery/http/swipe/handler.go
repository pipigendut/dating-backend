package swipe

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/delivery/http/response"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/services"
)

func NewSwipeHandler(r *gin.RouterGroup, swipeSvc services.SwipeService, storageService storageUsecase, authMiddleware gin.HandlerFunc) {
	handler := &SwipeHandler{
		swipeService:   swipeSvc,
		storageService: storageService,
	}

	swipeGroup := r.Group("/swipe")
	swipeGroup.Use(authMiddleware)
	{
		swipeGroup.GET("/candidates", handler.GetCandidates)
		swipeGroup.POST("/", handler.Swipe)
		swipeGroup.GET("/likes", handler.GetIncomingLikes)
		swipeGroup.GET("/likes/sent", handler.GetLikesSent)
		swipeGroup.POST("/unmatch/:entity_id", handler.Unmatch)
		swipeGroup.DELETE("/unlike/:entity_id", handler.Unlike)
		swipeGroup.GET("/likes/count", handler.GetLikesCount)
	}
}

type storageUsecase interface {
	GetPublicURL(key string) string
}

type SwipeHandler struct {
	swipeService   services.SwipeService
	storageService storageUsecase
}

// GetCandidates godoc
// @Summary      Get swipe candidates
// @Description  Fetches a list of potential entities (solo users or groups) based on entity type and user preferences.
// @Tags         swipe
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        query query SwipeCandidatesFilter false "Filter criteria"
// @Success      200  {object}  response.BaseResponse{data=[]response.EntityResponse} "Candidate list"
// @Router       /swipe/candidates [get]
func (h *SwipeHandler) GetCandidates(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var filter SwipeCandidatesFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid filter parameters", err.Error())
		return
	}

	svcFilter := services.SwipeFilter{
		Distance:          filter.Distance,
		MinAge:            filter.MinAge,
		MaxAge:            filter.MaxAge,
		Genders:           h.parseUUIDs(filter.Genders),
		Interests:         h.parseUUIDs(filter.Interests),
		RelationshipTypes: h.parseUUIDs(filter.RelationshipTypes),
		Latitude:          filter.Latitude,
		Longitude:         filter.Longitude,
		MinHeight:         filter.MinHeight,
		MaxHeight:         filter.MaxHeight,
	}

	// Set swiper entity
	swiperID, err := uuid.Parse(filter.SwiperEntityID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid swiper_entity_id", err.Error())
		return
	}
	svcFilter.SwiperEntityID = swiperID

	// Route to the correct service method based on entity_type
	var candidates []services.SwipeCandidate
	if filter.EntityType == "group" {
		if filter.EntityType != "" {
			et := entities.EntityType(filter.EntityType)
			svcFilter.EntityType = &et
		}
		candidates, err = h.swipeService.GetSwipeGroupCandidates(c.Request.Context(), userID, svcFilter, 10)
	} else {
		// Default: user swiper fetching user candidates
		candidates, err = h.swipeService.GetSwipeUserCandidates(c.Request.Context(), userID, svcFilter, 10)
	}

	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get swipe candidates", err.Error())
		return
	}

	respEntities := make([]response.EntityResponse, 0, len(candidates))
	for _, candidate := range candidates {
		resp := response.EntityResponse{
			ID:   candidate.Entity.ID,
			Type: string(candidate.Entity.Type),
		}

		switch candidate.Entity.Type {
		case entities.EntityTypeUser:
			if candidate.User != nil {
				userResp := response.ToUserResponse(candidate.User, h.storageService)
				resp.User = &userResp
			}

		case entities.EntityTypeGroup:
			if candidate.Group != nil {
				groupResp := h.buildGroupResponse(candidate.Group, h.storageService)
				resp.Group = &groupResp
			}
		}

		respEntities = append(respEntities, resp)
	}

	response.OK(c, respEntities)
}


// Swipe godoc
// @Summary      Record a swipe
// @Description  Records a user's swipe (LIKE, DISLIKE, or CRUSH) on a target entity and checks for a mutual match.
// @Tags         swipe
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body SwipeRequest true "Swipe action details"
// @Success      200  {object}  response.BaseResponse{data=MatchResponse} "Swipe recorded result"
// @Router       /swipe [post]
func (h *SwipeHandler) Swipe(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var req SwipeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	match, matchedEntity, err := h.swipeService.CreateSwipe(c.Request.Context(), userID, req.SwiperEntityID, req.SwipedEntityID, req.Direction)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error(), "Failed to record swipe")
		return
	}

	if match != nil {
		matchedEntResp := &response.EntityResponse{
			ID:   matchedEntity.ID,
			Type: string(matchedEntity.Type),
		}

		if matchedEntity.Type == entities.EntityTypeUser && matchedEntity.User != nil {
			ur := response.ToUserLiteResponse(matchedEntity.User, h.storageService)
			matchedEntResp.User = &ur
		} else if matchedEntity.Type == entities.EntityTypeGroup && matchedEntity.Group != nil {
			gr := h.buildGroupResponse(matchedEntity.Group, h.storageService)
			matchedEntResp.Group = &gr
		}

		response.OK(c, MatchResponse{
			IsMatch:       true,
			MatchID:       match.ID,
			MatchedEntity: matchedEntResp,
		})
		return
	}

	response.OK(c, MatchResponse{IsMatch: false})
}

// GetIncomingLikes godoc
// @Summary      Get incoming likes
// @Description  Returns a list of entities that have liked or crushed on the specified entity.
// @Tags         swipe
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        entity_id query string true "Entity ID to get likes for"
// @Success      200  {object}  response.BaseResponse{data=[]IncomingLikeResponse} "Incoming likes list"
// @Router       /swipe/likes [get]
func (h *SwipeHandler) GetIncomingLikes(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	likes, err := h.swipeService.GetIncomingLikes(c.Request.Context(), userID, 20, 0)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get incoming likes", err.Error())
		return
	}

	var respLikes []IncomingLikeResponse
	for _, like := range likes {
		entResp := response.EntityResponse{
			ID:   like.Entity.ID,
			Type: string(like.Entity.Type),
		}
		if like.User != nil {
			ur := response.ToUserResponse(like.User, h.storageService)
			entResp.User = &ur
		}
		if like.Group != nil {
			gr := h.buildGroupResponse(like.Group, h.storageService)
			entResp.Group = &gr
		}

		respLikes = append(respLikes, IncomingLikeResponse{
			Entity:         entResp,
			IsCrush:        like.IsCrush,
			IsBoosted:      like.IsBoosted,
			SwipeTime:      like.CreatedAt.Format(time.RFC3339),
			TargetEntityID: like.TargetEntityID.String(),
		})
	}

	response.OK(c, respLikes)
}

// GetLikesSent godoc
// @Summary      Get sent likes
// @Description  Returns a list of entities that the specified entity has liked.
// @Tags         swipe
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        entity_id query string true "Entity ID to get sent likes for"
// @Success      200  {object}  response.BaseResponse{data=[]SentLikeResponse} "Sent likes list"
// @Router       /swipe/likes/sent [get]
func (h *SwipeHandler) GetLikesSent(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	likes, err := h.swipeService.GetLikesSent(c.Request.Context(), userID, 20, 0)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get sent likes", err.Error())
		return
	}

	var respLikes []SentLikeResponse
	for _, like := range likes {
		entResp := response.EntityResponse{
			ID:   like.Entity.ID,
			Type: string(like.Entity.Type),
		}
		if like.User != nil {
			ur := response.ToUserResponse(like.User, h.storageService)
			entResp.User = &ur
		}
		if like.Group != nil {
			gr := h.buildGroupResponse(like.Group, h.storageService)
			entResp.Group = &gr
		}

		respLikes = append(respLikes, SentLikeResponse{
			Entity:         entResp,
			IsCrush:        like.IsCrush,
			IsBoosted:      like.IsBoosted,
			CreatedAt:      like.CreatedAt.Format(time.RFC3339),
			ExpiresAt:      like.ExpiresAt.Format(time.RFC3339),
			SwiperEntityID: like.SwiperEntityID.String(),
		})
	}

	response.OK(c, respLikes)
}

// Unmatch godoc
// @Summary      Unmatch an entity
// @Description  Removes an existing match between the specified swiper entity and another entity.
// @Tags         swipe
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        swiper_entity_id query string true "Swiper Entity ID"
// @Success      200  {object}  response.BaseResponse "Successfully unmatched"
// @Router       /swipe/unmatch/{entity_id} [post]
func (h *SwipeHandler) Unmatch(c *gin.Context) {
	targetEntityID, err := uuid.Parse(c.Param("entity_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid entity ID", err.Error())
		return
	}

	swiperEntityID, err := uuid.Parse(c.Query("swiper_entity_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid swiper_entity_id parameter", err.Error())
		return
	}

	if err := h.swipeService.DeleteMatch(c.Request.Context(), swiperEntityID, targetEntityID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to unmatch", err.Error())
		return
	}

	response.OK(c, nil)
}
 
// Unlike godoc
// @Summary      Unlike an entity
// @Description  Removes a one-way swipe from the current active entity to another entity.
// @Tags         swipe
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        entity_id path string true "Target Entity ID"
// @Param        swiper_entity_id query string true "Swiper Entity ID"
// @Success      200  {object}  response.BaseResponse "Successfully unliked"
// @Router       /swipe/unlike/{entity_id} [delete]
func (h *SwipeHandler) Unlike(c *gin.Context) {
	targetEntityID, err := uuid.Parse(c.Param("entity_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid entity ID", err.Error())
		return
	}
 
	swiperEntityID, err := uuid.Parse(c.Query("swiper_entity_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid swiper_entity_id parameter", err.Error())
		return
	}
 
	if err := h.swipeService.Unlike(c.Request.Context(), swiperEntityID, targetEntityID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to unlike", err.Error())
		return
	}
 
	response.OK(c, nil)
}

func (h *SwipeHandler) parseUUIDs(strs []string) []uuid.UUID {
	var uuids []uuid.UUID
	for _, s := range strs {
		if u, err := uuid.Parse(s); err == nil {
			uuids = append(uuids, u)
		}
	}
	return uuids
}

// GetLikesCount godoc
// @Summary      Get likes summary (count and last photo)
// @Description  Returns the total count of unexpired likes and the photo of the most recent liker for the specified entity.
// @Tags         swipe
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        entity_id query string true "Entity ID to get likes summary for"
// @Success      200  {object}  response.BaseResponse{data=LikesSummaryResponse} "Likes summary data"
// @Router       /swipe/likes/count [get]
func (h *SwipeHandler) GetLikesCount(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	summary, err := h.swipeService.GetLikesSummary(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get likes summary", err.Error())
		return
	}

	lastPhoto := summary.LastPhoto
	if lastPhoto != "" && !strings.HasPrefix(lastPhoto, "http") {
		lastPhoto = h.storageService.GetPublicURL(lastPhoto)
	}

	response.OK(c, LikesSummaryResponse{
		Count:     summary.Count,
		LastPhoto: lastPhoto,
	})
}

func (h *SwipeHandler) buildGroupResponse(g *entities.Group, storage storageUsecase) response.GroupResponse {
	resp := response.GroupResponse{
		ID:        g.ID,
		Name:      g.Name,
		CreatedBy: g.CreatedBy,
		Members:   make([]response.UserResponse, 0, len(g.Members)),
	}

	for _, m := range g.Members {
		if m.User != nil {
			userResp := response.ToUserResponse(m.User, storage)
			resp.Members = append(resp.Members, userResp)
			if userResp.MainPhoto != "" {
				resp.MainPhotos = append(resp.MainPhotos, userResp.MainPhoto)
			}
		}
	}

	return resp
}
