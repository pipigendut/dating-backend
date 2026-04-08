package bootstrap

import (
	handlerv1 "github.com/pipigendut/dating-backend/internal/delivery/http/handler/v1"
	v1 "github.com/pipigendut/dating-backend/router/api/v1"
)

// initHandlersV1 initializes all version 1 HTTP handlers
func initHandlersV1(infra *Infrastructure, repos *Repositories, svcs *Services) *v1.Handlers {
	return &v1.Handlers{
		User:         handlerv1.NewUserHandler(svcs.User, svcs.Storage, svcs.Verify, svcs.Entity),
		Auth:         handlerv1.NewAuthHandler(svcs.Auth, svcs.Storage),
		Master:       handlerv1.NewMasterHandler(svcs.Master, svcs.Ad),
		Monetization: handlerv1.NewMonetizationHandler(svcs.Subscription, repos.User, svcs.Storage),
		Swipe:        handlerv1.NewSwipeHandler(svcs.Swipe, svcs.Storage),
		Chat:         handlerv1.NewChatHandler(svcs.Chat, svcs.Swipe, svcs.Storage),
		Gif:          handlerv1.NewGifHandler(svcs.Gif, svcs.Chat),
		Admin:        handlerv1.NewAdminHandler(infra.DB, svcs.Config, svcs.Admin, repos.User, svcs.Storage),
		Entity:       handlerv1.NewEntityHandler(svcs.Entity, svcs.Storage),
		Group:        handlerv1.NewGroupHandler(svcs.Group, svcs.Storage),
		Device:       handlerv1.NewDeviceHandler(repos.Device, repos.Notification),
		Notification: handlerv1.NewNotificationHandler(svcs.NotificationConfig),
	}
}
