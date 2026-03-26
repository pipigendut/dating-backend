package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/pipigendut/dating-backend/internal/infra"
	"github.com/pipigendut/dating-backend/internal/infra/seeds"
)

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

	// 2. Run Master Seeders
	log.Println("Seeding master data...")
	if err := seeds.SeedMasterData(db); err != nil {
		log.Fatalf("Failed to execute master data seeders: %v", err)
	}
	log.Println("Seeding successful.")
}
