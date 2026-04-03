package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"github.com/pipigendut/dating-backend/internal/background"
	"github.com/pipigendut/dating-backend/internal/background/jobs"
	"github.com/pipigendut/dating-backend/internal/background/workers"
	"github.com/pipigendut/dating-backend/internal/chat/ws"
	"github.com/pipigendut/dating-backend/internal/delivery/http/admin"
	"github.com/pipigendut/dating-backend/internal/delivery/http/auth"
	"github.com/pipigendut/dating-backend/internal/delivery/http/chat"
	"github.com/pipigendut/dating-backend/internal/delivery/http/device"
	"github.com/pipigendut/dating-backend/internal/delivery/http/entity"
	"github.com/pipigendut/dating-backend/internal/delivery/http/groups"
	"github.com/pipigendut/dating-backend/internal/delivery/http/master"
	"github.com/pipigendut/dating-backend/internal/delivery/http/middleware"
	"github.com/pipigendut/dating-backend/internal/delivery/http/monetization"
	"github.com/pipigendut/dating-backend/internal/delivery/http/swipe"
	"github.com/pipigendut/dating-backend/internal/delivery/http/user"
	"github.com/pipigendut/dating-backend/internal/infra"
	"github.com/pipigendut/dating-backend/internal/infra/fcm"
	"github.com/pipigendut/dating-backend/internal/infra/ml"
	infraStorage "github.com/pipigendut/dating-backend/internal/infra/storage"
	"github.com/pipigendut/dating-backend/internal/providers/gif"
	"github.com/pipigendut/dating-backend/internal/providers/gif/klipy"
	"github.com/pipigendut/dating-backend/internal/repository"
	"github.com/pipigendut/dating-backend/internal/repository/impl"
	"github.com/pipigendut/dating-backend/internal/services"

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
	// 0. Load environment variables based on APP_ENV
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}
	envFile := ".env." + appEnv
	if err := godotenv.Load(envFile); err != nil {
		log.Printf("Warning: %s file not found, using system environment variables", envFile)
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

	// 1.1 Migrations and Seeding are now Manual (Check README.md)
	// Run: go run cmd/migrate/main.go up
	// Run: go run cmd/seed/main.go

	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisPass := os.Getenv("REDIS_PASSWORD")

	redisClient, err := infra.NewRedisClient(redisHost, redisPort, redisPass)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
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

	// 1.8 Setup Redis Repository & WebSocket Hub
	var (
		userRepo         repository.UserRepository
		sessionRepo      repository.SessionRepository
		masterRepo       repository.MasterRepository
		jobRepo          repository.JobRepository
		swipeRepo        repository.SwipeRepository
		subscriptionRepo repository.SubscriptionRepository
		redisRepo        repository.RedisRepository
		entityRepo       repository.EntityRepository
		groupRepo        repository.GroupRepository
		deviceRepo       repository.DeviceRepository
		notifRepo        repository.NotificationRepository
		adRepo           repository.AdvertisementRepository
	)

	redisRepo = impl.NewRedisRepository(redisClient)
	chatHub := ws.NewHub(redisRepo)
	go chatHub.Run(context.Background())
	if redisClient != nil {
		go chatHub.ListenToRedisPubSub(context.Background(), redisClient)
	}

	// 2. Initialize Layers
	userRepo = impl.NewUserRepo(db)
	sessionRepo = impl.NewSessionRepo(db)
	masterRepo = impl.NewMasterRepository(db)
	jobRepo = impl.NewJobRepository(db)
	swipeRepo = impl.NewSwipeRepository(db)
	subscriptionRepo = impl.NewSubscriptionRepository(db)
	entityRepo = impl.NewEntityRepository(db)
	groupRepo = impl.NewGroupRepository(db)
	deviceRepo = impl.NewDeviceRepository(db)
	notifRepo = impl.NewNotificationRepository(db)
	adRepo = impl.NewAdvertisementRepository(db)

	storageService := services.NewStorageService(storageImpl)

	// Background Jobs Initialization
	var asynqClient *asynq.Client
	var asynqServer *asynq.Server
	var notifySvc services.NotificationService
	var fcmClient *fcm.Client

	if redisClient != nil {
		redisOpt := asynq.RedisClientOpt{
			Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
			Password: redisPass,
		}

		asynqClient = asynq.NewClient(redisOpt)
		defer asynqClient.Close()

		asynqServer = asynq.NewServer(
			redisOpt,
			asynq.Config{
				Concurrency: 5,
				Queues: map[string]int{
					"default": 10,
				},
			},
		)

		notifySvc = services.NewNotificationService(asynqClient, redisRepo)

		// 1.10 Setup FCM Client
		svcAccountPath := os.Getenv("FIREBASE_SERVICE_ACCOUNT_PATH")
		if svcAccountPath != "" {
			var err error
			fcmClient, err = fcm.NewClient(svcAccountPath)
			if err != nil {
				log.Printf("Warning: FCM Client not initialized: %v", err)
			}
		}

		// Setup Worker with Tracking Router
		chatRepo := impl.NewChatRepository(db)
		notifyWorker := workers.NewNotificationWorker(chatRepo, userRepo, redisRepo, deviceRepo, groupRepo, notifRepo, fcmClient)
		cleanupHandler := jobs.NewUserCleanupHandler(db, storageImpl)

		jobRouter := background.NewJobRouter(jobRepo)
		jobRouter.RegisterHandler(services.TaskTypeNotificationGroup, notifyWorker.HandleNotificationGroupTask)
		jobRouter.RegisterHandler(services.TaskTypeNotificationMatch, notifyWorker.HandleMatchNotificationTask)
		jobRouter.RegisterHandler(services.TaskTypeNotificationLike, notifyWorker.HandleLikeNotificationTask)
		jobRouter.Mux().Handle(jobs.TaskUserCleanup, cleanupHandler)

		// Start Asynq Server non-blocking
		go func() {
			if err := asynqServer.Run(jobRouter.Mux()); err != nil {
				log.Fatalf("could not run asynq server: %v", err)
			}
		}()
	}

	configRepo := impl.NewConfigRepository(db)
	configSvc := services.NewConfigService(configRepo)

	promotionSvc := services.NewPromotionService(subscriptionRepo, userRepo, configSvc)
	userSvc := services.NewUserService(userRepo, jobRepo, sessionRepo, asynqClient, storageService, chatHub)
	authSvc := services.NewAuthService(userRepo, sessionRepo, entityRepo, storageService, promotionSvc)
	masterSvc := services.NewMasterService(masterRepo)
	adSvc := services.NewAdvertisementService(adRepo)

	verifySvc := services.NewVerificationService(userRepo, storageService, mlProvider, redisClient, configSvc)

	chatRepo := impl.NewChatRepository(db)
	chatSvc := services.NewChatService(chatRepo, userRepo, swipeRepo, redisRepo, notifySvc, chatHub)

	entitySvc := services.NewEntityService(entityRepo, userRepo)
	groupSvc := services.NewGroupService(groupRepo, entityRepo, userRepo)

	subscriptionService := services.NewSubscriptionService(subscriptionRepo, userRepo, redisRepo, configSvc)
	swipeSvc := services.NewSwipeService(db, configSvc, chatSvc, subscriptionService, notifySvc, swipeRepo, userRepo, entityRepo, groupRepo)
	adminSvc := services.NewAdminService(subscriptionRepo, userRepo)
	notifConfigSvc := services.NewNotificationConfigService(notifRepo)

	klipyAPIKey := os.Getenv("KLIPY_API_KEY")
	klipyShowAds := os.Getenv("KLIPY_SHOW_ADS") == "true"
	gifProviderName := os.Getenv("GIF_PROVIDER")

	var gifProvider gif.Provider
	if gifProviderName == "klipy" {
		gifProvider = klipy.NewKlipyProvider(klipyAPIKey, klipyShowAds)
	}
	// Add other providers here if needed

	gifSvc := services.NewGifService(gifProvider)

	r := gin.Default()

	// API Group
	v1 := r.Group("/api/v1")

	// Middleware setup
	authMiddleware := middleware.AuthMiddleware()
	if redisClient != nil {
		_ = middleware.NewCacheMiddleware(redisClient)
		acm := middleware.NewAntiCheatMiddleware(redisClient)
		_ = acm.RateLimitSwipe()
	}

	user.NewUserHandler(v1, userSvc, storageService, verifySvc, entitySvc, authMiddleware)
	auth.NewAuthHandler(v1, authSvc, storageService)
	master.NewMasterHandler(v1, masterSvc, adSvc)
	monetization.NewMonetizationHandler(v1, subscriptionService, userRepo, storageService, authMiddleware)
	swipe.NewSwipeHandler(v1, swipeSvc, storageService, authMiddleware)
	chat.NewChatHandler(v1, chatSvc, swipeSvc, storageService, authMiddleware)
	chat.NewGifHandler(v1, gifSvc, chatSvc, authMiddleware)
	admin.NewAdminHandler(db, configSvc, adminSvc, userRepo, storageService).RegisterRoutes(v1, authMiddleware)
	entity.NewEntityHandler(v1, entitySvc, storageService, authMiddleware)
	groupHandler := groups.NewGroupHandler(v1, groupSvc, storageService, authMiddleware)
	device.NewDeviceHandler(v1, deviceRepo, notifRepo, authMiddleware)
	user.NewNotificationHandler(v1, notifConfigSvc, authMiddleware)


	// Root routes for deep linking landing pages
	r.GET("/invite", groupHandler.HandleInviteRedirect)
	r.GET("/invite/:token", groupHandler.HandleInviteRedirect)

	// Well-known routes for Universal Links (served at root based on environment)
	r.Static("/.well-known", fmt.Sprintf("./well-known/%s", appEnv))

	// WebSocket Route
	v1.GET("/ws", func(c *gin.Context) {
		userIDStr := c.Query("user_id") // In production, this should come from JWT middleware
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid user_id"})
			return
		}
		ws.ServeWs(chatHub, chatSvc, c.Writer, c.Request, userID)
	})

	// Swagger route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Run HTTP Server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down servers...")

	// Graceful shutdown Asynq Server first
	if asynqServer != nil {
		asynqServer.Shutdown()
		log.Println("Asynq server stopped")
	}

	// Context with timeout to give HTTP server time to finish active requests
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}
