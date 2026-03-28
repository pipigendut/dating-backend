package infra

import (
	"github.com/hibiken/asynq"
	"log"
)

func NewAsynqClient(redisAddr string) *asynq.Client {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	log.Printf("[Asynq] Connected to Redis at %s", redisAddr)
	return client
}

func NewAsynqServer(redisAddr string, concurrency int) *asynq.Server {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: concurrency,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)
	return srv
}
