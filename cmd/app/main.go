package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pipigendut/dating-backend/internal/delivery/http/auth"
	"github.com/pipigendut/dating-backend/internal/delivery/http/middleware"
	"github.com/pipigendut/dating-backend/internal/delivery/http/user"
	"github.com/pipigendut/dating-backend/internal/infra"
	infraStorage "github.com/pipigendut/dating-backend/internal/infra/storage"
	"github.com/pipigendut/dating-backend/internal/repository/impl"
	"github.com/pipigendut/dating-backend/internal/usecases"

	_ "github.com/pipigendut/dating-backend/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Dating App API
// @version         1.0
// @description     This is a high-performance dating app backend server.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type 'Bearer ' followed by your JWT token.

func main() {
	// 0. Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

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

	// 1.1 Run Migrations
	if err := infra.Migrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisPass := os.Getenv("REDIS_PASSWORD")

	redisClient, err := infra.NewRedisClient(redisHost, redisPort, redisPass)
	if err != nil {
		log.Printf("Warning: Redis not connected: %v", err)
	}

	// 1.5 Setup S3 Client (Oracle Object Storage)
	s3AccessKey := os.Getenv("S3_ACCESS_KEY")
	s3SecretKey := os.Getenv("S3_SECRET_KEY")
	s3Endpoint := os.Getenv("S3_ENDPOINT")
	s3Region := os.Getenv("S3_REGION")
	if s3Region == "" {
		s3Region = "ap-singapore-1"
	}
	s3Bucket := os.Getenv("S3_BUCKET_NAME")

	s3Storage, err := infraStorage.NewS3Storage(s3AccessKey, s3SecretKey, s3Endpoint, s3Region, s3Bucket)
	if err != nil {
		log.Printf("Warning: S3 Storage not connected: %v", err)
	}

	// 2. Initialize Layers
	userRepo := impl.NewUserRepo(db)
	userUC := usecases.NewUserUsecase(userRepo)
	authUC := usecases.NewAuthUsecase(userRepo)
	storageUC := usecases.NewStorageUsecase(s3Storage)

	r := gin.Default()

	// API Group
	v1 := r.Group("/api/v1")

	// Middleware setup
	authMiddleware := middleware.AuthMiddleware()
	if redisClient != nil {
		_ = middleware.NewCacheMiddleware(redisClient)
		// Can use cacheMiddleware.Cache(time.Minute) on routes
	}

	user.NewUserHandler(v1, userUC, storageUC, authMiddleware)
	auth.NewAuthHandler(v1, authUC)

	// Swagger route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
