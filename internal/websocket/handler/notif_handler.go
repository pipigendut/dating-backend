package handler

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/websocket/hub"
)

// ServeNotifWs handles notification WebSocket connections.
func ServeNotifWs(h *hub.Hub, w http.ResponseWriter, r *http.Request, userID uuid.UUID) {
	conn, err := hub.Upgrade(w, r)
	if err != nil {
		log.Println("[WebSocket] notification upgrade error:", err)
		return
	}
	
	// Notifications might not need a chat service, so we pass nil or a specific notification service
	hub.RegisterClient(h, conn, nil, userID)
}
