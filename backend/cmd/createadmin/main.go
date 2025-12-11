package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/Wei-Shaw/sub2api/backend/config"
	postgressRepo "github.com/Wei-Shaw/sub2api/backend/internal/repository/postgres"
	"github.com/Wei-Shaw/sub2api/backend/internal/service/admin"
	"github.com/Wei-Shaw/sub2api/backend/pkg/logger"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: createadmin <username> <password>")
		fmt.Println("Example: createadmin admin admin123")
		os.Exit(1)
	}

	username := os.Args[1]
	password := os.Args[2]

	// Load config
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logCfg := logger.LoggingConfig{
		Level:           cfg.Logging.Level,
		Format:          cfg.Logging.Format,
		OutputPath:      cfg.Logging.OutputPath,
		ErrorOutputPath: cfg.Logging.ErrorOutputPath,
	}
	zapLogger, err := logger.NewLogger(logCfg)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Connect to database
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Database,
		cfg.Database.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Create admin service
	adminRepo := postgressRepo.NewAdminRepository(db)
	adminService := admin.NewAdminService(adminRepo, cfg.Security.JWTSecret, cfg.Security.TokenExpiration, zapLogger)

	// Create admin
	ctx := context.Background()
	createdAdmin, err := adminService.CreateAdmin(ctx, username, password)
	if err != nil {
		log.Fatalf("Failed to create admin: %v", err)
	}

	fmt.Printf("Admin created successfully:\n")
	fmt.Printf("  ID:       %d\n", createdAdmin.ID)
	fmt.Printf("  Username: %s\n", createdAdmin.Username)
	fmt.Printf("  Email:    %s\n", createdAdmin.Email)
	fmt.Printf("  Active:   %t\n", createdAdmin.IsActive)
	fmt.Printf("\nYou can now log in with username: %s and the password you provided.\n", username)
}
