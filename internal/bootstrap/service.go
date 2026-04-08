package bootstrap

import (
	"os"

	"github.com/hibiken/asynq"
	"github.com/pipigendut/dating-backend/internal/chat/ws"
	"github.com/pipigendut/dating-backend/internal/providers/gif"
	"github.com/pipigendut/dating-backend/internal/providers/gif/klipy"
	"github.com/pipigendut/dating-backend/internal/services"
)

// Services holds all service instances
type Services struct {
	User               *services.UserService
	Auth               *services.AuthService
	Master             *services.MasterService
	Ad                 services.AdvertisementService
	Verify             *services.VerificationService
	Chat               services.ChatService
	Entity             services.EntityService
	Group              services.GroupService
	Subscription       services.SubscriptionService
	Swipe              services.SwipeService
	Admin              services.AdminService
	NotificationConfig services.NotificationConfigService
	Gif                services.GifService
	Storage            *services.StorageService
	Notification       services.NotificationService
	Config             services.ConfigService
}

// initServices initializes all service layer implementations
func initServices(infra *Infrastructure, repos *Repositories, asynqClient *asynq.Client, chatHub *ws.Hub) *Services {
	storageService := services.NewStorageService(infra.Storage)
	configSvc := services.NewConfigService(repos.Config)

	// Notification service depends on asynq client
	var notifySvc services.NotificationService
	if asynqClient != nil {
		notifySvc = services.NewNotificationService(asynqClient, repos.Redis)
	}

	promotionSvc := services.NewPromotionService(repos.Subscription, repos.User, configSvc)
	userSvc := services.NewUserService(repos.User, repos.Job, repos.Session, asynqClient, storageService, chatHub)
	authSvc := services.NewAuthService(repos.User, repos.Session, repos.Entity, storageService, promotionSvc)
	masterSvc := services.NewMasterService(repos.Master)
	adSvc := services.NewAdvertisementService(repos.Ad)

	verifySvc := services.NewVerificationService(repos.User, storageService, infra.MLProvider, infra.Redis, configSvc)

	chatSvc := services.NewChatService(repos.Chat, repos.User, repos.Swipe, repos.Redis, notifySvc, chatHub)

	entitySvc := services.NewEntityService(repos.Entity, repos.User)
	groupSvc := services.NewGroupService(repos.Group, repos.Entity, repos.User)

	subscriptionService := services.NewSubscriptionService(repos.Subscription, repos.User, repos.Redis, configSvc)
	swipeSvc := services.NewSwipeService(infra.DB, configSvc, chatSvc, subscriptionService, notifySvc, repos.Swipe, repos.User, repos.Entity, repos.Group)
	adminSvc := services.NewAdminService(repos.Subscription, repos.User)
	notifConfigSvc := services.NewNotificationConfigService(repos.Notification)

	// Gif Service Setup
	klipyAPIKey := os.Getenv("KLIPY_API_KEY")
	klipyShowAds := os.Getenv("KLIPY_SHOW_ADS") == "true"
	gifProviderName := os.Getenv("GIF_PROVIDER")

	var gifProvider gif.Provider
	if gifProviderName == "klipy" {
		gifProvider = klipy.NewKlipyProvider(klipyAPIKey, klipyShowAds)
	}

	gifSvc := services.NewGifService(gifProvider)

	return &Services{
		User:               userSvc,
		Auth:               authSvc,
		Master:             masterSvc,
		Ad:                 adSvc,
		Verify:             verifySvc,
		Chat:               chatSvc,
		Entity:             entitySvc,
		Group:              groupSvc,
		Subscription:       subscriptionService,
		Swipe:              swipeSvc,
		Admin:              adminSvc,
		NotificationConfig: notifConfigSvc,
		Gif:                gifSvc,
		Storage:            storageService,
		Notification:       notifySvc,
		Config:             configSvc,
	}
}
