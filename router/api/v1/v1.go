package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/pipigendut/dating-backend/internal/chat/ws"
	handlerv1 "github.com/pipigendut/dating-backend/internal/delivery/http/handler/v1"
	"github.com/pipigendut/dating-backend/internal/services"
)

type Handlers struct {
	Auth         *handlerv1.AuthHandler
	User         *handlerv1.UserHandler
	Notification *handlerv1.NotificationHandler
	Master       *handlerv1.MasterHandler
	Monetization *handlerv1.MonetizationHandler
	Swipe        *handlerv1.SwipeHandler
	Chat         *handlerv1.ChatHandler
	Gif          *handlerv1.GifHandler
	Admin        *handlerv1.AdminHandler
	Entity       *handlerv1.EntityHandler
	Group        *handlerv1.GroupHandler
	Device       *handlerv1.DeviceHandler
}

type RouterConfig struct {
	AppEnv         string
	ChatHub        *ws.Hub
	ChatService    services.ChatService
	AuthMiddleware gin.HandlerFunc
}

func RegisterRoutes(r *gin.Engine, h *Handlers, cfg RouterConfig) {
	v1 := r.Group("/api/v1")

	registerAuthRoutes(v1, h, cfg)
	registerMasterRoutes(v1, h, cfg)
	registerUserRoutes(v1, h, cfg)
	registerMonetizationRoutes(v1, h, cfg)
	registerSwipeRoutes(v1, h, cfg)
	registerChatRoutes(v1, h, cfg)
	registerEntityRoutes(v1, h, cfg)
	registerGroupRoutes(v1, h, cfg)
	registerDeviceRoutes(v1, h, cfg)
	registerAdminRoutes(v1, h, cfg)
}
