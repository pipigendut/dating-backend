package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	// 0. Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// 1. Setup Database Connection String
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")

	if dbHost == "" || dbUser == "" || dbName == "" {
		log.Fatal("Database configuration missing in .env")
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPass, dbHost, dbPort, dbName)

	// 2. Parse Commands
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		log.Fatal("Please specify a command: up, down, or force <version>")
	}

	command := args[0]

	// 3. Initialize Migrate
	m, err := migrate.New(
		"file://migrations",
		dsn,
	)
	if err != nil {
		log.Fatalf("Could not initialize migrate: %v", err)
	}

	// 4. Execute Commands
	switch command {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migrate up failed: %v", err)
		}
		log.Println("Migrate up successful.")
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migrate down failed: %v", err)
		}
		log.Println("Migrate down successful.")
	case "force":
		if len(args) < 2 {
			log.Fatal("Force command requires a version number")
		}
		var version int
		if _, err := fmt.Sscanf(args[1], "%d", &version); err != nil {
			log.Fatalf("Invalid version number: %v", err)
		}
		if err := m.Force(version); err != nil {
			log.Fatalf("Migrate force failed: %v", err)
		}
		log.Println("Migrate force successful.")
	case "step":
		if len(args) < 2 {
			log.Fatal("Step command requires a number (e.g., 1 or -1)")
		}
		var n int
		if _, err := fmt.Sscanf(args[1], "%d", &n); err != nil {
			log.Fatalf("Invalid step number: %v", err)
		}
		if err := m.Steps(n); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migrate step failed: %v", err)
		}
		log.Printf("Migrate step %d successful.\n", n)
	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("Could not get version: %v", err)
		}
		fmt.Printf("Version: %d, Dirty: %v\n", version, dirty)
	default:
		log.Fatalf("Unknown command: %s", command)
	}
}
