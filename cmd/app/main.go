package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pipigendut/dating-backend/internal/delivery/http/admin"
	"github.com/pipigendut/dating-backend/internal/delivery/http/auth"
	"github.com/pipigendut/dating-backend/internal/delivery/http/chat"
	"github.com/pipigendut/dating-backend/internal/delivery/http/master"
	"github.com/pipigendut/dating-backend/internal/delivery/http/middleware"
	"github.com/pipigendut/dating-backend/internal/delivery/http/monetization"
	"github.com/pipigendut/dating-backend/internal/delivery/http/swipe"
	"github.com/pipigendut/dating-backend/internal/delivery/http/user"
	"github.com/pipigendut/dating-backend/internal/delivery/ws"
	"github.com/pipigendut/dating-backend/internal/infra"
	"github.com/pipigendut/dating-backend/internal/infra/kafka"
	"github.com/pipigendut/dating-backend/internal/infra/ml"
	"github.com/pipigendut/dating-backend/internal/infra/seeds"
	infraStorage "github.com/pipigendut/dating-backend/internal/infra/storage"
	"github.com/pipigendut/dating-backend/internal/repository/impl"
	"github.com/pipigendut/dating-backend/internal/services"
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

	// 1.2 Run Master Seeders
	if err := seeds.SeedMasterData(db); err != nil {
		log.Fatalf("Failed to execute master data seeders: %v", err)
	}

	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisPass := os.Getenv("REDIS_PASSWORD")

	redisClient, err := infra.NewRedisClient(redisHost, redisPort, redisPass)
	if err != nil {
		log.Printf("Warning: Redis not connected: %v", err)
	}

	// 1.5 Setup Storage Provider
	storageProvider := os.Getenv("STORAGE_PROVIDER")
	s3AccessKey := os.Getenv("S3_ACCESS_KEY")
	s3SecretKey := os.Getenv("S3_SECRET_KEY")
	s3Endpoint := os.Getenv("S3_ENDPOINT")
	s3Region := os.Getenv("S3_REGION")
	s3Bucket := os.Getenv("S3_BUCKET_NAME")

	if storageProvider == "AWS" {
		// AWS doesn't need custom endpoint usually, it uses regional ones
		s3Endpoint = ""
	}

	if s3Region == "" {
		s3Region = "ap-singapore-1"
	}

	var storageImpl infraStorage.StorageProvider
	var errS3 error

	storageImpl, errS3 = infraStorage.NewS3Storage(s3AccessKey, s3SecretKey, s3Endpoint, s3Region, s3Bucket)
	if errS3 != nil {
		log.Printf("Warning: Storage Provider (%s) not connected: %v", storageProvider, errS3)
	}

	mlProvider, err := ml.NewProvider()
	if err != nil {
		log.Printf("Warning: Face Verification Provider not initialized: %v", err)
	} else if mlProvider != nil {
		defer mlProvider.Close()
	}

	// 1.7 Setup Kafka
	kafkaBrokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	if os.Getenv("KAFKA_BROKERS") == "" {
		kafkaBrokers = []string{"localhost:9092"}
	}
	kafkaProducer := kafka.NewProducer(kafkaBrokers)
	defer kafkaProducer.Close()

	// 1.8 Setup WebSocket Manager
	wsManager := ws.NewManager()
	go wsManager.Run()

	// 2. Initialize Layers
	userRepo := impl.NewUserRepo(db)
	sessionRepo := impl.NewSessionRepo(db)
	masterRepo := impl.NewMasterRepository(db)

	storageUC := usecases.NewStorageUsecase(storageImpl)
	userUC := usecases.NewUserUsecase(userRepo, storageUC)
	authUC := usecases.NewAuthUsecase(userRepo, sessionRepo, storageUC)
	masterUC := usecases.NewMasterUsecase(masterRepo)

	configRepo := impl.NewConfigRepository(db)
	configSvc := services.NewConfigService(configRepo)

	verifyUC := usecases.NewVerificationService(userRepo, storageUC, mlProvider, redisClient, configSvc)

	chatRepo := impl.NewChatRepository(db)
	chatSvc := services.NewChatService(chatRepo, kafkaProducer)

	swipeRepo := impl.NewSwipeRepository(db)
	subscriptionRepo := impl.NewSubscriptionRepository(db)

	subscriptionService := services.NewSubscriptionService(subscriptionRepo, userRepo)
	swipeSvc := services.NewSwipeService(db, configSvc, chatSvc, subscriptionService, swipeRepo)
	adminSvc := services.NewAdminService(subscriptionRepo, userRepo)

	// 2.1 Kafka Consumer
	chatConsumer := kafka.NewConsumer(kafkaBrokers, "chat-group", "chat.messages", chatRepo, wsManager)
	go chatConsumer.Start(context.Background())
	defer chatConsumer.Close()

	r := gin.Default()

	// API Group
	v1 := r.Group("/api/v1")

	// Middleware setup
	authMiddleware := middleware.AuthMiddleware()
	var anticheatMiddleware gin.HandlerFunc
	if redisClient != nil {
		_ = middleware.NewCacheMiddleware(redisClient)
		acm := middleware.NewAntiCheatMiddleware(redisClient)
		anticheatMiddleware = acm.RateLimitSwipe()
	}

	user.NewUserHandler(v1, userUC, storageUC, verifyUC, authMiddleware)
	auth.NewAuthHandler(v1, authUC)
	master.NewMasterHandler(v1, masterUC)
	monetization.NewMonetizationHandler(v1, subscriptionService, userRepo, authMiddleware)
	swipe.NewSwipeHandler(v1, swipeSvc, storageUC, authMiddleware, anticheatMiddleware)
	chat.NewChatHandler(v1, chatSvc, storageUC, authMiddleware)
	admin.NewAdminHandler(db, configSvc, adminSvc, userRepo).RegisterRoutes(v1, authMiddleware)

	// WebSocket Route
	wsHandler := ws.NewHandler(wsManager, chatSvc)
	v1.GET("/ws", wsHandler.HandleWebSocket)

	// Swagger route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
