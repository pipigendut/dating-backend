package chat

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/delivery/http/response"
	"github.com/pipigendut/dating-backend/internal/services"
	"github.com/pipigendut/dating-backend/internal/usecases"
)

type ChatHandler struct {
	chatService services.ChatService
	storageUC   *usecases.StorageUsecase
}

func NewChatHandler(r *gin.RouterGroup, chatService services.ChatService, storageUC *usecases.StorageUsecase, authMiddleware gin.HandlerFunc) {
	handler := &ChatHandler{
		chatService: chatService,
		storageUC:   storageUC,
	}

	chatGroup := r.Group("/chat")
	chatGroup.Use(authMiddleware)
	{
		chatGroup.GET("/conversations", handler.GetConversations)
		chatGroup.GET("/conversations/:id/messages", handler.GetMessages)
		chatGroup.POST("/conversations/match/:target_user_id", handler.GetOrCreateMatchConversation)
		chatGroup.GET("/upload-url", handler.GetUploadURL)
	}
}

// GetConversations godoc
// @Summary List user conversations
// @Description Get a list of all active conversations for the authenticated user
// @Tags Chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.BaseResponse{data=[]ConversationResponse}
// @Router /chat/conversations [get]
func (h *ChatHandler) GetConversations(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	convs, err := h.chatService.GetConversations(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get conversations", err.Error())
		return
	}

	resp := make([]ConversationResponse, len(convs))
	for i, conv := range convs {
		unreadCount, _ := h.chatService.GetUnreadCount(c.Request.Context(), conv.ID, userID)
		resp[i] = ToConversationResponse(&conv, userID, unreadCount, h.storageUC)
	}

	response.OK(c, resp)
}

// GetMessages godoc
// @Summary Get conversation messages
// @Description Get paginated message history for a specific conversation
// @Tags Chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conversation ID"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.BaseResponse{data=[]MessageResponse}
// @Router /chat/conversations/{id}/messages [get]
func (h *ChatHandler) GetMessages(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	convID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid conversation ID", err.Error())
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	msgs, err := h.chatService.GetMessages(c.Request.Context(), convID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get messages", err.Error())
		return
	}

	resp := make([]MessageResponse, len(msgs))
	for i, msg := range msgs {
		resp[i] = ToMessageResponse(&msg, userID)
	}

	response.OK(c, resp)
}

// GetOrCreateMatchConversation godoc
// @Summary Initialize conversation with a match
// @Description Get or create a 1:1 conversation between matched users
// @Tags Chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param target_user_id path string true "Target User ID"
// @Success 200 {object} response.BaseResponse{data=ConversationResponse}
// @Router /chat/conversations/match/{target_user_id} [post]
func (h *ChatHandler) GetOrCreateMatchConversation(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	targetUserID, err := uuid.Parse(c.Param("target_user_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid target user ID", err.Error())
		return
	}

	conv, err := h.chatService.GetOrCreateConversation(c.Request.Context(), userID, targetUserID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to manage conversation", err.Error())
		return
	}

	unreadCount, _ := h.chatService.GetUnreadCount(c.Request.Context(), conv.ID, userID)
	response.OK(c, ToConversationResponse(conv, userID, unreadCount, h.storageUC))
}

// GetUploadURL godoc
// @Summary Get presigned upload URL for chat media
// @Description Generate a presigned S3/Oracle URL for uploading chat attachments
// @Tags Chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param conversation_id query string true "Conversation ID"
// @Success 200 {object} response.BaseResponse{data=ChatUploadURLResponse}
// @Router /chat/upload-url [get]
func (h *ChatHandler) GetUploadURL(c *gin.Context) {
	convID, err := uuid.Parse(c.Query("conversation_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid conversation ID", err.Error())
		return
	}

	url, key, err := h.storageUC.GetChatUploadURL(c.Request.Context(), convID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to generate upload URL", err.Error())
		return
	}

	response.OK(c, ChatUploadURLResponse{
		UploadURL: url,
		FileKey:   key,
	})
}
