package v1

import (
	"github.com/gin-gonic/gin"
	// WebSocket Routes
	"github.com/pipigendut/dating-backend/internal/websocket"
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

	// Register WebSocket routes (/ws, /ws/chat, /ws/notif, /ws/presence)
	websocket.RegisterRoutes(v1, cfg.ChatHub, cfg.ChatService)
}
