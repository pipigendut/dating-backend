package handler

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/websocket/hub"
)

// ServePresenceWs handles presence tracking WebSocket connections.
func ServePresenceWs(h *hub.Hub, w http.ResponseWriter, r *http.Request, userID uuid.UUID) {
	conn, err := hub.Upgrade(w, r)
	if err != nil {
		log.Println("[WebSocket] presence upgrade error:", err)
		return
	}
	
	// Presence tracking only needs the hub registration
	hub.RegisterClient(h, conn, nil, userID)
}
