package handler

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/pipigendut/dating-backend/internal/websocket/hub"
)

// ServeWs upgrades an HTTP connection to a WebSocket connection and
// registers the new client with the Hub.
// This is the entry point called from the HTTP router.
func ServeWs(h *hub.Hub, chatSvc hub.ChatServiceInterface, w http.ResponseWriter, r *http.Request, userID uuid.UUID) {
	conn, err := hub.Upgrade(w, r)
	if err != nil {
		log.Println("[WebSocket] upgrade error:", err)
		return
	}
	hub.RegisterClient(h, conn, chatSvc, userID)
}
