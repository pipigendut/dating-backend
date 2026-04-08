package websocket

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/websocket/handler"
	"github.com/pipigendut/dating-backend/internal/websocket/hub"
)

// RegisterRoutes registers all WebSocket routes.
func RegisterRoutes(rg *gin.RouterGroup, h *hub.Hub, chatSvc hub.ChatServiceInterface) {
	// Main entry point (chat) - legacy /ws path for compatibility
	rg.GET("/ws", handleWs(h, chatSvc, handler.ServeWs))

	// Explicitly named WS endpoints
	wsGroup := rg.Group("/ws")
	{
		wsGroup.GET("/chat", handleWs(h, chatSvc, handler.ServeWs))
		wsGroup.GET("/notif", func(c *gin.Context) {
			userID, err := getUserID(c)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			handler.ServeNotifWs(h, c.Writer, c.Request, userID)
		})
		wsGroup.GET("/presence", func(c *gin.Context) {
			userID, err := getUserID(c)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			handler.ServePresenceWs(h, c.Writer, c.Request, userID)
		})
	}
}

func handleWs(h *hub.Hub, chatSvc hub.ChatServiceInterface, srv func(*hub.Hub, hub.ChatServiceInterface, http.ResponseWriter, *http.Request, uuid.UUID)) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := getUserID(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		srv(h, chatSvc, c.Writer, c.Request, userID)
	}
}

func getUserID(c *gin.Context) (uuid.UUID, error) {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		return uuid.Nil, http.ErrNoLocation // generic error or custom
	}
	return uuid.Parse(userIDStr)
}
