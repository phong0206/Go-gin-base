package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"os"
	"time"

	"github.com/phong0206/go_gin_restful/internal/common/helper"
	"github.com/phong0206/go_gin_restful/internal/config"
	"github.com/phong0206/go_gin_restful/pkg/database"
)

func main() {
	log.Println("[SEED] Starting admin seeder...")

	// 1. Load application config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[SEED] Failed to load config: %v", err)
	}

	// 2. Connect to database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.OpenDatabase(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("[SEED] Database connection failed: %v", err)
	}
	defer db.Close()

	// 3. Admin seed information
	adminEmail := getEnv("SEED_ADMIN_EMAIL", "admin@example.com")
	adminUsername := getEnv("SEED_ADMIN_USERNAME", "admin")
	adminPassword := getEnv("SEED_ADMIN_PASSWORD", "Admin@123456")

	// 4. Check if admin already exists
	var existingID int
	checkQuery := `SELECT id FROM users WHERE email = $1 LIMIT 1`
	err = db.GetContext(ctx, &existingID, checkQuery, adminEmail)
	if err == nil {
		log.Printf("[SEED] Admin user (%s) already exists with ID: %d. Skipping creation.", adminEmail, existingID)
		return
	}
	if !errors.Is(err, sql.ErrNoRows) {
		log.Fatalf("[SEED] Failed to check existing user: %v", err)
	}

	// 5. Hash password
	hasher := helper.NewPasswordHelper()
	passwordHash, err := hasher.Hash(adminPassword)
	if err != nil {
		log.Fatalf("[SEED] Failed to hash password: %v", err)
	}

	// 6. Insert admin user
	insertQuery := `
		INSERT INTO users (email, username, password_hash, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, true, NOW(), NOW())
		RETURNING id
	`
	var newID int
	if err := db.QueryRowContext(ctx, insertQuery, adminEmail, adminUsername, passwordHash).Scan(&newID); err != nil {
		log.Fatalf("[SEED] Failed to insert admin user: %v", err)
	}

	log.Println("==================================================")
	log.Printf("[SEED] Successfully seeded admin user!")
	log.Printf("  - ID:       %d", newID)
	log.Printf("  - Username: %s", adminUsername)
	log.Printf("  - Email:    %s", adminEmail)
	log.Printf("  - Password: %s", adminPassword)
	log.Println("==================================================")
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
