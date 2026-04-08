package bootstrap

import (
	"github.com/pipigendut/dating-backend/internal/repository"
	"github.com/pipigendut/dating-backend/internal/repository/impl"
)

// Repositories holds all repository instances
type Repositories struct {
	User         repository.UserRepository
	Session      repository.SessionRepository
	Master       repository.MasterRepository
	Job          repository.JobRepository
	Swipe        repository.SwipeRepository
	Subscription repository.SubscriptionRepository
	Redis        repository.RedisRepository
	Entity       repository.EntityRepository
	Group        repository.GroupRepository
	Device       repository.DeviceRepository
	Notification repository.NotificationRepository
	Ad           repository.AdvertisementRepository
	Config       repository.ConfigRepository
	Chat         repository.ChatRepository
}

// initRepositories initializes all repository implementations
func initRepositories(infra *Infrastructure) *Repositories {
	return &Repositories{
		User:         impl.NewUserRepo(infra.DB),
		Session:      impl.NewSessionRepo(infra.DB),
		Master:       impl.NewMasterRepository(infra.DB),
		Job:          impl.NewJobRepository(infra.DB),
		Swipe:        impl.NewSwipeRepository(infra.DB),
		Subscription: impl.NewSubscriptionRepository(infra.DB),
		Redis:        impl.NewRedisRepository(infra.Redis),
		Entity:       impl.NewEntityRepository(infra.DB),
		Group:        impl.NewGroupRepository(infra.DB),
		Device:       impl.NewDeviceRepository(infra.DB),
		Notification: impl.NewNotificationRepository(infra.DB),
		Ad:           impl.NewAdvertisementRepository(infra.DB),
		Config:       impl.NewConfigRepository(infra.DB),
		Chat:         impl.NewChatRepository(infra.DB),
	}
}
