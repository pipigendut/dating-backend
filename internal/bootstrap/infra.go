package bootstrap

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/pipigendut/dating-backend/internal/infra"
	"github.com/pipigendut/dating-backend/internal/infra/fcm"
	"github.com/pipigendut/dating-backend/internal/infra/ml"
	infraStorage "github.com/pipigendut/dating-backend/internal/infra/storage"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Infrastructure holds all core infrastructure dependencies
type Infrastructure struct {
	DB          *gorm.DB
	Redis       *redis.Client
	Storage     infraStorage.StorageProvider
	MLProvider  ml.FaceVerificationProvider
	FCMClient   *fcm.Client
	AppEnv      string
	RedisHost   string
	RedisPort   string
	RedisPass   string
}

// initInfra initializes all core infrastructure components
func initInfra() *Infrastructure {
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
		s3Endpoint = ""
	}
	if s3Region == "" {
		s3Region = "ap-singapore-1"
	}

	storageImpl, errS3 := infraStorage.NewS3Storage(s3AccessKey, s3SecretKey, s3Endpoint, s3Region, s3Bucket)
	if errS3 != nil {
		log.Printf("Warning: Storage Provider (%s) not connected: %v", storageProvider, errS3)
	}

	mlProvider, err := ml.NewProvider()
	if err != nil {
		log.Printf("Warning: Face Verification Provider not initialized: %v", err)
	}

	// 1.10 Setup FCM Client
	var fcmClient *fcm.Client
	svcAccountPath := os.Getenv("FIREBASE_SERVICE_ACCOUNT_PATH")
	if svcAccountPath != "" {
		var err error
		fcmClient, err = fcm.NewClient(svcAccountPath)
		if err != nil {
			log.Printf("Warning: FCM Client not initialized: %v", err)
		}
	}

	return &Infrastructure{
		DB:         db,
		Redis:      redisClient,
		Storage:    storageImpl,
		MLProvider: mlProvider,
		FCMClient:  fcmClient,
		AppEnv:     appEnv,
		RedisHost:  redisHost,
		RedisPort:  redisPort,
		RedisPass:  redisPass,
	}
}
