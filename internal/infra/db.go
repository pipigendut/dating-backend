package infra

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"github.com/pipigendut/dating-backend/internal/entities"
)

type Config struct {
	DBHost     string
	DBUser     string
	DBPassword string
	DBName     string
	DBPort     string
	DBSSLMode  string
}

func NewPostgresDB(cfg Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode)

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,        // Enable color
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, err
	}

	// Fix: Drop incorrect unique index for messages if it exists
	db.Migrator().DropIndex(&entities.Message{}, "idx_conv_created")
 
	// Auto Migration
	err = db.AutoMigrate(
		// Core
		&entities.Entity{},
		&entities.User{},
		&entities.Group{},
		&entities.GroupMember{},
		// Interactions
		&entities.Swipe{},
		&entities.Match{},
		&entities.EntityBoost{},
		&entities.EntityImpression{},
		&entities.EntityUnmatch{},
		// Chat
		&entities.Conversation{},
		&entities.ConversationParticipant{},
		&entities.Message{},
		&entities.MessageRead{},
		&entities.GroupInvite{},
		&entities.UserPresence{},
		// User Profile
		&entities.AuthProvider{},
		&entities.Photo{},
		&entities.Device{},
		&entities.RefreshToken{},
		// Master Tables
		&entities.MasterGender{},
		&entities.MasterRelationshipType{},
		&entities.MasterInterest{},
		&entities.MasterLanguage{},
		// Pivot Tables (many-to-many)
		&entities.UserInterestedGender{},
		&entities.UserInterest{},
		&entities.UserLanguage{},
		// Config & Monetization
		&entities.AppConfig{},
		&entities.SubscriptionPlan{},
		&entities.SubscriptionPrice{},
		&entities.SubscriptionPlanFeature{},
		&entities.UserSubscription{},
		&entities.UserConsumable{},
		&entities.ConsumablePackage{},
		// Background Jobs
		&entities.Job{},
		// Notifications
		&entities.NotificationSetting{},
		&entities.UserNotificationSetting{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to auto migrate: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Connection Pool Settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}
