package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/chat/ws"
)

func registerChatRoutes(v1 *gin.RouterGroup, h *Handlers, cfg RouterConfig) {
	chat := v1.Group("/chat")
	chat.Use(cfg.AuthMiddleware)
	{
		chat.GET("/conversations", h.Chat.GetConversations)
		chat.GET("/new-matches", h.Chat.GetNewMatches)
		chat.GET("/conversations/:id/messages", h.Chat.GetMessages)
		chat.GET("/conversations/match/:matchId", h.Chat.GetConversationByMatch)
		chat.GET("/upload-url", h.Chat.GetUploadURL)

		// Gifs
		gifs := chat.Group("/gifs")
		{
			gifs.GET("/search", h.Gif.Search)
			gifs.GET("/trending", h.Gif.Trending)
			gifs.POST("/send", h.Gif.SendGifMessage)
		}
	}

	// WebSocket Route (Public entry, auth inside or via query)
	v1.GET("/ws", func(c *gin.Context) {
		userIDStr := c.Query("user_id")
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
			return
		}
		ws.ServeWs(cfg.ChatHub, cfg.ChatService, c.Writer, c.Request, userID)
	})
}
