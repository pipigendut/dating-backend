package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/pipigendut/dating-backend/internal/delivery/http/middleware"
	"github.com/pipigendut/dating-backend/internal/delivery/http/user"
	"github.com/pipigendut/dating-backend/internal/infra"
	"github.com/pipigendut/dating-backend/internal/repository"
	"github.com/pipigendut/dating-backend/internal/repository/impl"
	"github.com/pipigendut/dating-backend/internal/usecases"
)

func main() {
	// 1. Setup Infrastructure
	dbCfg := infra.Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		DBPort:     os.Getenv("DB_PORT"),
		DBSSLMode:  "disable",
	}

	db, err := infra.NewPostgresDB(dbCfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisPass := os.Getenv("REDIS_PASSWORD")

	redisClient, err := infra.NewRedisClient(redisHost, redisPort, redisPass)
	if err != nil {
		log.Printf("Warning: Redis not connected: %v", err)
	}

	// 2. Initialize Layers
	userRepo := impl.NewUserRepo(db)
	userUC := usecases.NewUserUsecase(userRepo)
	
	r := gin.Default()
	
	// API Group
	v1 := r.Group("/api/v1")
	
	// Middleware setup
	if redisClient != nil {
		_ = middleware.NewCacheMiddleware(redisClient)
		// Can use cacheMiddleware.Cache(time.Minute) on routes
	}

	user.NewUserHandler(v1, userUC)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
