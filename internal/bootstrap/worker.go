package bootstrap

import (
	"fmt"
	"log"

	"github.com/hibiken/asynq"
	"github.com/pipigendut/dating-backend/internal/background"
	"github.com/pipigendut/dating-backend/internal/background/jobs"
	"github.com/pipigendut/dating-backend/internal/background/workers"
	"github.com/pipigendut/dating-backend/internal/services"
)

// initWorkers initializes the Asynq client, server, and task handlers
func initWorkers(infra *Infrastructure, repos *Repositories) (*asynq.Client, *asynq.Server) {
	if infra.Redis == nil {
		return nil, nil
	}

	redisOpt := asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%s", infra.RedisHost, infra.RedisPort),
		Password: infra.RedisPass,
	}

	asynqClient := asynq.NewClient(redisOpt)

	asynqServer := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: 5,
			Queues: map[string]int{
				"default": 10,
			},
		},
	)

	// Setup Worker with Tracking Router
	notifyWorker := workers.NewNotificationWorker(repos.Chat, repos.User, repos.Redis, repos.Device, repos.Group, repos.Notification, infra.FCMClient)
	cleanupHandler := jobs.NewUserCleanupHandler(infra.DB, infra.Storage)

	jobRouter := background.NewJobRouter(repos.Job)
	jobRouter.RegisterHandler(services.TaskTypeNotificationGroup, notifyWorker.HandleNotificationGroupTask)
	jobRouter.RegisterHandler(services.TaskTypeNotificationMatch, notifyWorker.HandleMatchNotificationTask)
	jobRouter.RegisterHandler(services.TaskTypeNotificationLike, notifyWorker.HandleLikeNotificationTask)
	jobRouter.Mux().Handle(jobs.TaskUserCleanup, cleanupHandler)

	// Start Asynq Server in background
	go func() {
		if err := asynqServer.Run(jobRouter.Mux()); err != nil {
			log.Fatalf("could not run asynq server: %v", err)
		}
	}()

	return asynqClient, asynqServer
}
