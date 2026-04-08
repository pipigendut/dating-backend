package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	base "github.com/pipigendut/dating-backend/internal/delivery/http/dto"
	"github.com/pipigendut/dating-backend/internal/entities"
	"github.com/pipigendut/dating-backend/internal/services"
)

type GifHandler struct {
	gifService  services.GifService
	chatService services.ChatService
}

func NewGifHandler(gifSvc services.GifService, chatSvc services.ChatService) *GifHandler {
	return &GifHandler{
		gifService:  gifSvc,
		chatService: chatSvc,
	}
}

func (h *GifHandler) Search(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	query := c.Query("q")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	locale := c.DefaultQuery("locale", "id")

	gifs, err := h.gifService.Search(c.Request.Context(), userID.String(), query, locale, limit, offset)
	if err != nil {
		base.Error(c, http.StatusInternalServerError, "Failed to search GIFs", err.Error())
		return
	}

	base.OK(c, gifs)
}

func (h *GifHandler) Trending(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	locale := c.DefaultQuery("locale", "id")

	gifs, err := h.gifService.Trending(c.Request.Context(), userID.String(), locale, limit, offset)
	if err != nil {
		base.Error(c, http.StatusInternalServerError, "Failed to fetch trending GIFs", err.Error())
		return
	}

	base.OK(c, gifs)
}

type SendGifRequest struct {
	ConversationID uuid.UUID `json:"conversation_id" binding:"required"`
	URL            string    `json:"url" binding:"required"`
	Width          int       `json:"width"`
	Height         int       `json:"height"`
}

func (h *GifHandler) SendGifMessage(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var req SendGifRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// For Gif messages, Content is the URL. We should use standard SendMessage
	// wait, SendMessage doesn't accept a Metadata payload out of the box in the `ChatService` interface.
	// But it does accept `messageType`. Let's use `SendMessage`!
	// (Note: To properly save Width/Height, we might need to modify `SendMessage` to take metadata,
	//  but we can stick to what's available or JSON encode into Content, or simply save the URL as content).

	err := h.chatService.SendMessage(c.Request.Context(), userID, req.ConversationID, entities.MessageTypeGif, req.URL, &entities.MessageMetadata{
		GifProvider: "klipy", // Can be based on config later
		ImageWidth:  req.Width,
		ImageHeight: req.Height,
	})
	if err != nil {
		base.Error(c, http.StatusInternalServerError, "Failed to send GIF message", err.Error())
		return
	}

	base.OK(c, gin.H{"status": "queued"})
}
