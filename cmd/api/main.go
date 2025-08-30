package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mem_bank/configs"
	"mem_bank/internal/app"
	"mem_bank/pkg/database"
	"mem_bank/pkg/logger"
)

func main() {
	// Load configuration
	config, err := configs.LoadConfig("")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	appLogger, err := logger.NewLogger(&config.Logging)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Initialize database
	db, err := database.NewGormConnection(&config.Database)
	if err != nil {
		appLogger.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	appLogger.Info("Database connection established")

	// Initialize Redis
	redisConfig := database.RedisConfig{
		Host:     config.Redis.Host,
		Port:     config.Redis.Port,
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
	}
	redisClient, err := database.NewRedisClient(redisConfig, appLogger)
	if err != nil {
		appLogger.WithError(err).Fatal("Failed to connect to Redis")
	}
	defer redisClient.Close()

	// Create application config
	appConfig := app.Config{
		Host:         config.Server.Host,
		Port:         config.Server.Port,
		Mode:         config.Server.Mode,
		ReadTimeout:  config.Server.ReadTimeout,
		WriteTimeout: config.Server.WriteTimeout,
		IdleTimeout:  config.Server.IdleTimeout,
	}

	// Create and start application
	application := app.NewApp(db.DB, redisClient, appLogger, config)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in goroutine
	go func() {
		if err := application.Start(ctx, appConfig); err != nil {
			appLogger.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := application.Shutdown(shutdownCtx); err != nil {
		appLogger.WithError(err).Error("Server forced to shutdown")
	}

	appLogger.Info("Server exiting")
}
